import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ConfigModule } from '@nestjs/config';
import { WinstonModule } from 'nest-winston';
import { APP_FILTER, APP_INTERCEPTOR, APP_PIPE } from '@nestjs/core';
import { AuthModule } from './auth/auth.module';
import { User } from './entities/user.entity';
import { HealthController } from '../../../shared/nestjs/controllers/health.controller';
import { HttpExceptionFilter } from '../../../shared/nestjs/filters/http-exception.filter';
import { LoggingInterceptor } from '../../../shared/nestjs/interceptors/logging.interceptor';
import { ResponseInterceptor } from '../../../shared/nestjs/interceptors/response.interceptor';
import { CustomValidationPipe } from '../../../shared/nestjs/pipes/validation.pipe';
import { winstonConfig } from '../../../shared/nestjs/config/winston.config';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    WinstonModule.forRoot(winstonConfig),
    TypeOrmModule.forRoot({
      type: 'postgres',
      url: process.env.DATABASE_URL,
      entities: [User],
      synchronize: true, // Don't use in production
      logging: process.env.NODE_ENV === 'development',
    }),
    AuthModule,
  ],
  controllers: [HealthController],
  providers: [
    {
      provide: APP_FILTER,
      useClass: HttpExceptionFilter,
    },
    {
      provide: APP_INTERCEPTOR,
      useClass: LoggingInterceptor,
    },
    {
      provide: APP_INTERCEPTOR,
      useClass: ResponseInterceptor,
    },
    {
      provide: APP_PIPE,
      useClass: CustomValidationPipe,
    },
  ],
})
export class AppModule {}

