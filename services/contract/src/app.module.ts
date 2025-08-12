import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ConfigModule } from '@nestjs/config';
import { ContractModule } from './contract/contract.module';
import { Contract } from './entities/contract.entity';
import { HealthController } from './controllers/health.controller';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    TypeOrmModule.forRoot({
      type: 'postgres',
      url: process.env.DATABASE_URL,
      entities: [Contract],
      synchronize: true, // Don't use in production
      logging: process.env.NODE_ENV === 'development',
    }),
    ContractModule,
  ],
  controllers: [HealthController],
  providers: [],
})
export class AppModule {}

