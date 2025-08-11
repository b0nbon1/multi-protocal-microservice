import { Controller, Get } from '@nestjs/common';
import { HealthCheck, ServiceStatus } from '../../types/common.types';

@Controller('health')
export class HealthController {
  private readonly startTime = Date.now();

  @Get()
  check(): HealthCheck {
    const uptime = Date.now() - this.startTime;
    
    return {
      service: process.env.SERVICE_NAME || 'unknown-service',
      status: ServiceStatus.HEALTHY,
      timestamp: new Date(),
      uptime: Math.floor(uptime / 1000), // in seconds
      version: process.env.SERVICE_VERSION || '1.0.0',
    };
  }
}

