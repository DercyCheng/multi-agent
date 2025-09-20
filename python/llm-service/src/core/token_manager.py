"""
Token Budget Manager for Multi-Agent LLM Service
Advanced cost tracking and budget enforcement with multi-tenant support
"""

import asyncio
import logging
import time
from typing import Dict, Optional, List, Tuple
from dataclasses import dataclass, field
from decimal import Decimal, ROUND_HALF_UP
import json

import asyncpg
import redis.asyncio as redis

from src.config import TokenConfig
from src.core.models import TokenBudget, CompletionRequest, CompletionResponse, Usage

logger = logging.getLogger(__name__)

@dataclass
class CostCalculation:
    """Cost calculation result"""
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    cost_usd: Decimal
    model: str
    provider: str

@dataclass
class BudgetAlert:
    """Budget alert information"""
    user_id: str
    tenant_id: str
    alert_type: str  # warning, limit_reached, exceeded
    current_usage: Decimal
    budget_limit: Decimal
    threshold_percentage: float
    timestamp: float = field(default_factory=time.time)

class TokenManager:
    """Advanced token budget management with cost optimization"""
    
    def __init__(self, config: TokenConfig):
        self.config = config
        self.db_pool: Optional[asyncpg.Pool] = None
        self.redis_client: Optional[redis.Redis] = None
        self.cost_per_token: Dict[str, Decimal] = {
            model: Decimal(str(cost)) 
            for model, cost in config.cost_per_token.items()
        }
        self.budget_cache: Dict[str, TokenBudget] = {}
        self.usage_cache: Dict[str, Decimal] = {}
        self._lock = asyncio.Lock()
        
        # Alert thresholds
        self.alert_thresholds = [0.5, 0.8, 0.9, 0.95, 1.0]  # 50%, 80%, 90%, 95%, 100%
        
    async def initialize(self):
        """Initialize token manager"""
        logger.info("Initializing token manager")
        
        # Initialize database connection
        await self._init_database()
        
        # Initialize Redis connection
        await self._init_redis()
        
        # Load budget data
        await self._load_budget_cache()
        
        # Start background tasks
        asyncio.create_task(self._budget_sync_task())
        asyncio.create_task(self._usage_aggregation_task())
        asyncio.create_task(self._budget_reset_task())
        
        logger.info("Token manager initialized successfully")
    
    async def _init_database(self):
        """Initialize PostgreSQL connection"""
        try:
            self.db_pool = await asyncpg.create_pool(
                host="localhost",
                port=5432,
                database="multiagent",
                user="postgres",
                password="password",
                min_size=5,
                max_size=20
            )
            logger.info("Token manager database connection established")
        except Exception as e:
            logger.error(f"Failed to initialize database: {e}")
            raise
    
    async def _init_redis(self):
        """Initialize Redis connection"""
        try:
            self.redis_client = redis.from_url(
                "redis://localhost:6379/2",  # Use database 2 for token management
                decode_responses=True
            )
            await self.redis_client.ping()
            logger.info("Token manager Redis connection established")
        except Exception as e:
            logger.error(f"Failed to initialize Redis: {e}")
            raise
    
    async def _load_budget_cache(self):
        """Load budget data into cache"""
        if not self.db_pool:
            return
        
        try:
            async with self.db_pool.acquire() as conn:
                rows = await conn.fetch("""
                    SELECT user_id, tenant_id, total_budget, used_budget, 
                           daily_limit, monthly_limit, last_reset
                    FROM token_budgets
                    WHERE is_active = true
                """)
                
                for row in rows:
                    budget = TokenBudget(
                        user_id=row['user_id'],
                        tenant_id=row['tenant_id'],
                        total_budget=float(row['total_budget']),
                        used_budget=float(row['used_budget']),
                        remaining_budget=float(row['total_budget']) - float(row['used_budget']),
                        daily_limit=float(row['daily_limit']) if row['daily_limit'] else None,
                        monthly_limit=float(row['monthly_limit']) if row['monthly_limit'] else None,
                        last_reset=int(row['last_reset'].timestamp()) if row['last_reset'] else int(time.time())
                    )
                    
                    cache_key = f"{budget.tenant_id}:{budget.user_id}"
                    self.budget_cache[cache_key] = budget
                
                logger.info(f"Loaded {len(self.budget_cache)} budget entries into cache")
                
        except Exception as e:
            logger.error(f"Failed to load budget cache: {e}")
    
    async def check_budget_availability(
        self, 
        user_id: str, 
        tenant_id: str, 
        estimated_cost: Decimal
    ) -> Tuple[bool, Optional[str]]:
        """Check if user has sufficient budget for the request"""
        
        budget = await self.get_user_budget(user_id, tenant_id)
        
        if not budget:
            return False, "No budget found for user"
        
        # Check total budget
        if budget.remaining_budget < float(estimated_cost):
            return False, f"Insufficient budget. Required: ${estimated_cost:.4f}, Available: ${budget.remaining_budget:.4f}"
        
        # Check daily limit
        if budget.daily_limit:
            daily_usage = await self._get_daily_usage(user_id, tenant_id)
            if daily_usage + estimated_cost > Decimal(str(budget.daily_limit)):
                return False, f"Daily limit exceeded. Limit: ${budget.daily_limit:.2f}"
        
        # Check monthly limit
        if budget.monthly_limit:
            monthly_usage = await self._get_monthly_usage(user_id, tenant_id)
            if monthly_usage + estimated_cost > Decimal(str(budget.monthly_limit)):
                return False, f"Monthly limit exceeded. Limit: ${budget.monthly_limit:.2f}"
        
        return True, None
    
    async def reserve_budget(
        self, 
        user_id: str, 
        tenant_id: str, 
        estimated_cost: Decimal,
        request_id: str
    ) -> bool:
        """Reserve budget for a request"""
        
        if not self.config.budget_enforcement_enabled:
            return True
        
        async with self._lock:
            # Check availability
            available, reason = await self.check_budget_availability(user_id, tenant_id, estimated_cost)
            
            if not available:
                logger.warning(f"Budget reservation failed for {user_id}: {reason}")
                return False
            
            # Reserve in Redis with expiration
            reservation_key = f"budget_reservation:{tenant_id}:{user_id}:{request_id}"
            await self.redis_client.setex(
                reservation_key, 
                300,  # 5 minutes expiration
                str(estimated_cost)
            )
            
            # Update cache
            cache_key = f"{tenant_id}:{user_id}"
            if cache_key in self.budget_cache:
                self.budget_cache[cache_key].remaining_budget -= float(estimated_cost)
            
            logger.debug(f"Reserved ${estimated_cost:.4f} for user {user_id}")
            return True
    
    async def consume_budget(
        self, 
        user_id: str, 
        tenant_id: str, 
        actual_cost: Decimal,
        request_id: str,
        usage: Usage,
        model: str,
        provider: str
    ) -> bool:
        """Consume actual budget after request completion"""
        
        if not self.config.cost_tracking_enabled:
            return True
        
        async with self._lock:
            try:
                # Release reservation
                reservation_key = f"budget_reservation:{tenant_id}:{user_id}:{request_id}"
                reserved_amount = await self.redis_client.get(reservation_key)
                if reserved_amount:
                    await self.redis_client.delete(reservation_key)
                    reserved_amount = Decimal(reserved_amount)
                else:
                    reserved_amount = Decimal('0')
                
                # Record usage in database
                if self.db_pool:
                    async with self.db_pool.acquire() as conn:
                        await conn.execute("""
                            INSERT INTO token_usage (
                                user_id, tenant_id, request_id, model, provider,
                                prompt_tokens, completion_tokens, total_tokens,
                                cost_usd, created_at
                            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
                        """, 
                            user_id, tenant_id, request_id, model, provider,
                            usage.prompt_tokens, usage.completion_tokens, usage.total_tokens,
                            float(actual_cost)
                        )
                        
                        # Update budget
                        await conn.execute("""
                            UPDATE token_budgets 
                            SET used_budget = used_budget + $1,
                                updated_at = NOW()
                            WHERE user_id = $2 AND tenant_id = $3
                        """, float(actual_cost), user_id, tenant_id)
                
                # Update cache
                cache_key = f"{tenant_id}:{user_id}"
                if cache_key in self.budget_cache:
                    budget = self.budget_cache[cache_key]
                    
                    # Adjust for difference between reserved and actual
                    adjustment = actual_cost - reserved_amount
                    budget.used_budget += float(actual_cost)
                    budget.remaining_budget = budget.total_budget - budget.used_budget
                    
                    # Check for alerts
                    await self._check_budget_alerts(budget)
                
                # Update usage cache for quick access
                usage_key = f"usage:{tenant_id}:{user_id}"
                current_usage = self.usage_cache.get(usage_key, Decimal('0'))
                self.usage_cache[usage_key] = current_usage + actual_cost
                
                logger.debug(f"Consumed ${actual_cost:.4f} for user {user_id}")
                return True
                
            except Exception as e:
                logger.error(f"Failed to consume budget: {e}")
                return False
    
    async def release_budget(
        self, 
        user_id: str, 
        tenant_id: str, 
        request_id: str
    ):
        """Release reserved budget if request fails"""
        
        reservation_key = f"budget_reservation:{tenant_id}:{user_id}:{request_id}"
        reserved_amount = await self.redis_client.get(reservation_key)
        
        if reserved_amount:
            reserved_amount = Decimal(reserved_amount)
            await self.redis_client.delete(reservation_key)
            
            # Update cache
            cache_key = f"{tenant_id}:{user_id}"
            if cache_key in self.budget_cache:
                self.budget_cache[cache_key].remaining_budget += float(reserved_amount)
            
            logger.debug(f"Released ${reserved_amount:.4f} for user {user_id}")
    
    async def calculate_cost(
        self, 
        model: str, 
        provider: str, 
        usage: Usage
    ) -> CostCalculation:
        """Calculate cost for token usage"""
        
        # Get cost per token for the model
        cost_per_token = self.cost_per_token.get(model, Decimal('0.002'))  # Default fallback
        
        # Different pricing for prompt vs completion tokens (some models)
        prompt_cost_multiplier = Decimal('1.0')
        completion_cost_multiplier = Decimal('1.0')
        
        # Adjust for specific models with different pricing tiers
        if 'gpt-4' in model.lower():
            prompt_cost_multiplier = Decimal('1.0')
            completion_cost_multiplier = Decimal('2.0')  # Output tokens often cost more
        elif 'claude-3-opus' in model.lower():
            prompt_cost_multiplier = Decimal('1.0')
            completion_cost_multiplier = Decimal('3.0')
        
        # Calculate costs
        prompt_cost = (Decimal(str(usage.prompt_tokens)) * cost_per_token * prompt_cost_multiplier) / 1000
        completion_cost = (Decimal(str(usage.completion_tokens)) * cost_per_token * completion_cost_multiplier) / 1000
        total_cost = prompt_cost + completion_cost
        
        # Round to 6 decimal places
        total_cost = total_cost.quantize(Decimal('0.000001'), rounding=ROUND_HALF_UP)
        
        return CostCalculation(
            prompt_tokens=usage.prompt_tokens,
            completion_tokens=usage.completion_tokens,
            total_tokens=usage.total_tokens,
            cost_usd=total_cost,
            model=model,
            provider=provider
        )
    
    async def estimate_cost(
        self, 
        request: CompletionRequest, 
        model: str, 
        provider: str
    ) -> Decimal:
        """Estimate cost for a completion request"""
        
        # Estimate token count
        estimated_prompt_tokens = self._estimate_prompt_tokens(request)
        estimated_completion_tokens = request.max_tokens or 500
        
        # Create mock usage for calculation
        mock_usage = Usage(
            prompt_tokens=estimated_prompt_tokens,
            completion_tokens=estimated_completion_tokens,
            total_tokens=estimated_prompt_tokens + estimated_completion_tokens
        )
        
        cost_calc = await self.calculate_cost(model, provider, mock_usage)
        return cost_calc.cost_usd
    
    def _estimate_prompt_tokens(self, request: CompletionRequest) -> int:
        """Estimate prompt token count"""
        total_chars = 0
        
        for message in request.messages:
            total_chars += len(message.content)
            if message.name:
                total_chars += len(message.name)
        
        # Add overhead for message formatting
        total_chars += len(request.messages) * 10
        
        # Add tool definitions if present
        if request.tools:
            for tool in request.tools:
                total_chars += len(json.dumps(tool.dict()))
        
        # Rough conversion: 4 characters per token
        return total_chars // 4
    
    async def get_user_budget(self, user_id: str, tenant_id: str) -> Optional[TokenBudget]:
        """Get user budget information"""
        
        cache_key = f"{tenant_id}:{user_id}"
        
        # Check cache first
        if cache_key in self.budget_cache:
            return self.budget_cache[cache_key]
        
        # Load from database
        if not self.db_pool:
            return None
        
        try:
            async with self.db_pool.acquire() as conn:
                row = await conn.fetchrow("""
                    SELECT user_id, tenant_id, total_budget, used_budget,
                           daily_limit, monthly_limit, last_reset
                    FROM token_budgets
                    WHERE user_id = $1 AND tenant_id = $2 AND is_active = true
                """, user_id, tenant_id)
                
                if not row:
                    # Create default budget
                    return await self._create_default_budget(user_id, tenant_id)
                
                budget = TokenBudget(
                    user_id=row['user_id'],
                    tenant_id=row['tenant_id'],
                    total_budget=float(row['total_budget']),
                    used_budget=float(row['used_budget']),
                    remaining_budget=float(row['total_budget']) - float(row['used_budget']),
                    daily_limit=float(row['daily_limit']) if row['daily_limit'] else None,
                    monthly_limit=float(row['monthly_limit']) if row['monthly_limit'] else None,
                    last_reset=int(row['last_reset'].timestamp()) if row['last_reset'] else int(time.time())
                )
                
                # Cache the result
                self.budget_cache[cache_key] = budget
                return budget
                
        except Exception as e:
            logger.error(f"Failed to get user budget: {e}")
            return None
    
    async def _create_default_budget(self, user_id: str, tenant_id: str) -> TokenBudget:
        """Create default budget for new user"""
        
        default_budget = self.config.default_budget
        
        if self.db_pool:
            try:
                async with self.db_pool.acquire() as conn:
                    await conn.execute("""
                        INSERT INTO token_budgets (
                            user_id, tenant_id, total_budget, used_budget,
                            daily_limit, monthly_limit, is_active, created_at, updated_at
                        ) VALUES ($1, $2, $3, 0, NULL, NULL, true, NOW(), NOW())
                    """, user_id, tenant_id, default_budget)
            except Exception as e:
                logger.error(f"Failed to create default budget: {e}")
        
        budget = TokenBudget(
            user_id=user_id,
            tenant_id=tenant_id,
            total_budget=default_budget,
            used_budget=0.0,
            remaining_budget=default_budget
        )
        
        cache_key = f"{tenant_id}:{user_id}"
        self.budget_cache[cache_key] = budget
        
        return budget
    
    async def _get_daily_usage(self, user_id: str, tenant_id: str) -> Decimal:
        """Get daily usage for user"""
        
        if not self.db_pool:
            return Decimal('0')
        
        try:
            async with self.db_pool.acquire() as conn:
                result = await conn.fetchval("""
                    SELECT COALESCE(SUM(cost_usd), 0)
                    FROM token_usage
                    WHERE user_id = $1 AND tenant_id = $2
                    AND created_at >= CURRENT_DATE
                """, user_id, tenant_id)
                
                return Decimal(str(result or 0))
                
        except Exception as e:
            logger.error(f"Failed to get daily usage: {e}")
            return Decimal('0')
    
    async def _get_monthly_usage(self, user_id: str, tenant_id: str) -> Decimal:
        """Get monthly usage for user"""
        
        if not self.db_pool:
            return Decimal('0')
        
        try:
            async with self.db_pool.acquire() as conn:
                result = await conn.fetchval("""
                    SELECT COALESCE(SUM(cost_usd), 0)
                    FROM token_usage
                    WHERE user_id = $1 AND tenant_id = $2
                    AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
                """, user_id, tenant_id)
                
                return Decimal(str(result or 0))
                
        except Exception as e:
            logger.error(f"Failed to get monthly usage: {e}")
            return Decimal('0')
    
    async def _check_budget_alerts(self, budget: TokenBudget):
        """Check and send budget alerts"""
        
        utilization = budget.budget_utilization / 100.0
        
        for threshold in self.alert_thresholds:
            if utilization >= threshold:
                alert_type = "exceeded" if threshold >= 1.0 else "warning" if threshold >= 0.9 else "limit_reached"
                
                alert = BudgetAlert(
                    user_id=budget.user_id,
                    tenant_id=budget.tenant_id,
                    alert_type=alert_type,
                    current_usage=Decimal(str(budget.used_budget)),
                    budget_limit=Decimal(str(budget.total_budget)),
                    threshold_percentage=threshold * 100
                )
                
                await self._send_budget_alert(alert)
                break  # Send only the highest threshold alert
    
    async def _send_budget_alert(self, alert: BudgetAlert):
        """Send budget alert (implement notification logic)"""
        
        # Store alert in database
        if self.db_pool:
            try:
                async with self.db_pool.acquire() as conn:
                    await conn.execute("""
                        INSERT INTO budget_alerts (
                            user_id, tenant_id, alert_type, current_usage,
                            budget_limit, threshold_percentage, created_at
                        ) VALUES ($1, $2, $3, $4, $5, $6, NOW())
                    """, 
                        alert.user_id, alert.tenant_id, alert.alert_type,
                        float(alert.current_usage), float(alert.budget_limit),
                        alert.threshold_percentage
                    )
            except Exception as e:
                logger.error(f"Failed to store budget alert: {e}")
        
        # Log alert
        logger.warning(
            f"Budget alert for user {alert.user_id}: {alert.alert_type} "
            f"({alert.threshold_percentage:.1f}% threshold reached)"
        )
        
        # TODO: Implement actual notification (email, webhook, etc.)
    
    async def _budget_sync_task(self):
        """Background task to sync budget cache with database"""
        while True:
            try:
                await asyncio.sleep(300)  # Sync every 5 minutes
                
                if self.db_pool:
                    # Reload budget cache
                    await self._load_budget_cache()
                
                logger.debug("Budget cache synchronized")
                
            except Exception as e:
                logger.error(f"Budget sync task error: {e}")
    
    async def _usage_aggregation_task(self):
        """Background task to aggregate usage statistics"""
        while True:
            try:
                await asyncio.sleep(3600)  # Run every hour
                
                if self.db_pool:
                    async with self.db_pool.acquire() as conn:
                        # Aggregate hourly usage
                        await conn.execute("""
                            INSERT INTO usage_aggregates (
                                tenant_id, user_id, model, provider,
                                period_start, period_end, period_type,
                                total_requests, total_tokens, total_cost,
                                created_at
                            )
                            SELECT 
                                tenant_id, user_id, model, provider,
                                DATE_TRUNC('hour', created_at) as period_start,
                                DATE_TRUNC('hour', created_at) + INTERVAL '1 hour' as period_end,
                                'hourly' as period_type,
                                COUNT(*) as total_requests,
                                SUM(total_tokens) as total_tokens,
                                SUM(cost_usd) as total_cost,
                                NOW() as created_at
                            FROM token_usage
                            WHERE created_at >= NOW() - INTERVAL '2 hours'
                            AND created_at < DATE_TRUNC('hour', NOW())
                            GROUP BY tenant_id, user_id, model, provider, DATE_TRUNC('hour', created_at)
                            ON CONFLICT (tenant_id, user_id, model, provider, period_start, period_type) 
                            DO NOTHING
                        """)
                
                logger.debug("Usage aggregation completed")
                
            except Exception as e:
                logger.error(f"Usage aggregation task error: {e}")
    
    async def _budget_reset_task(self):
        """Background task to reset budgets based on schedule"""
        while True:
            try:
                await asyncio.sleep(86400)  # Check daily
                
                current_time = time.time()
                
                # Reset daily budgets at midnight
                # Reset monthly budgets on first day of month
                # This is a simplified implementation
                
                if self.db_pool:
                    async with self.db_pool.acquire() as conn:
                        # Reset daily budgets (simplified - would need proper timezone handling)
                        await conn.execute("""
                            UPDATE token_budgets 
                            SET used_budget = 0, updated_at = NOW()
                            WHERE daily_limit IS NOT NULL
                            AND last_reset < CURRENT_DATE
                        """)
                
                logger.debug("Budget reset check completed")
                
            except Exception as e:
                logger.error(f"Budget reset task error: {e}")
    
    async def get_usage_statistics(
        self, 
        user_id: str, 
        tenant_id: str, 
        period: str = "day"
    ) -> Dict[str, any]:
        """Get usage statistics for user"""
        
        if not self.db_pool:
            return {}
        
        try:
            async with self.db_pool.acquire() as conn:
                if period == "day":
                    interval = "1 day"
                elif period == "week":
                    interval = "7 days"
                elif period == "month":
                    interval = "30 days"
                else:
                    interval = "1 day"
                
                stats = await conn.fetchrow("""
                    SELECT 
                        COUNT(*) as total_requests,
                        SUM(total_tokens) as total_tokens,
                        SUM(cost_usd) as total_cost,
                        AVG(cost_usd) as avg_cost_per_request,
                        COUNT(DISTINCT model) as unique_models
                    FROM token_usage
                    WHERE user_id = $1 AND tenant_id = $2
                    AND created_at >= NOW() - INTERVAL %s
                """ % f"'{interval}'", user_id, tenant_id)
                
                return {
                    "period": period,
                    "total_requests": stats['total_requests'] or 0,
                    "total_tokens": stats['total_tokens'] or 0,
                    "total_cost_usd": float(stats['total_cost'] or 0),
                    "avg_cost_per_request": float(stats['avg_cost_per_request'] or 0),
                    "unique_models": stats['unique_models'] or 0,
                }
                
        except Exception as e:
            logger.error(f"Failed to get usage statistics: {e}")
            return {}
    
    async def shutdown(self):
        """Shutdown token manager"""
        logger.info("Shutting down token manager")
        
        if self.db_pool:
            await self.db_pool.close()
        
        if self.redis_client:
            await self.redis_client.close()
        
        logger.info("Token manager shutdown completed")