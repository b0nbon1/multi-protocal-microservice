import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { PassportModule } from '@nestjs/passport';
import { JwtModule } from '@nestjs/jwt';
import { DisputeService } from './dispute.service';
import { DisputeController } from './dispute.controller';
import { Dispute } from '../entities/dispute.entity';
import { JwtStrategy } from '../auth/jwt.strategy';

@Module({
  imports: [
    TypeOrmModule.forFeature([Dispute]),
    PassportModule,
    JwtModule.register({
      secret: process.env.JWT_SECRET || 'your-secret-key',
      signOptions: { expiresIn: '24h' },
    }),
  ],
  controllers: [DisputeController],
  providers: [DisputeService, JwtStrategy],
  exports: [DisputeService],
})
export class DisputeModule {}

