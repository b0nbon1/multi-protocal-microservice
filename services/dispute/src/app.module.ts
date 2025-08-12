import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ConfigModule } from '@nestjs/config';
import { DisputeModule } from './dispute/dispute.module';
import { Dispute } from './entities/dispute.entity';
import { HealthController } from './controllers/health.controller';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    TypeOrmModule.forRoot({
      type: 'postgres',
      url: process.env.DATABASE_URL,
      entities: [Dispute],
      synchronize: true, // Don't use in production
      logging: process.env.NODE_ENV === 'development',
    }),
    DisputeModule,
  ],
  controllers: [HealthController],
  providers: [],
})
export class AppModule {}