import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ConfigModule } from '@nestjs/config';
import { PaymentModule } from './payment/payment.module';
import { Wallet } from './entities/wallet.entity';
import { Transaction } from './entities/transaction.entity';
import { HealthController } from './controllers/health.controller';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    TypeOrmModule.forRoot({
      type: 'postgres',
      url: process.env.DATABASE_URL,
      entities: [Wallet, Transaction],
      synchronize: true, // Don't use in production
      logging: process.env.NODE_ENV === 'development',
    }),
    PaymentModule,
  ],
  controllers: [HealthController],
  providers: [],
})
export class AppModule {}