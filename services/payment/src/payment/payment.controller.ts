import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  UseGuards,
} from '@nestjs/common';
import { PaymentService } from './payment.service';
import { DepositDto, TransferDto } from '../dto/payment.dto';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@Controller()
@UseGuards(JwtAuthGuard)
export class PaymentController {
  constructor(private readonly paymentService: PaymentService) {}

  @Get('wallets/:userId')
  getWalletBalance(@Param('userId') userId: string) {
    return this.paymentService.getWalletBalance(userId);
  }

  @Post('wallets/deposit')
  deposit(@Body() depositDto: DepositDto) {
    return this.paymentService.deposit(depositDto);
  }

  @Post('transfers')
  transfer(@Body() transferDto: TransferDto) {
    return this.paymentService.transfer(transferDto);
  }

  @Get('transactions/:userId')
  getTransactionHistory(@Param('userId') userId: string) {
    return this.paymentService.getTransactionHistory(userId);
  }
}

