import { IsString, IsUUID, IsNumber, IsEnum, IsOptional, Min } from 'class-validator';
import { TransactionType } from '../../../../shared/types/common.types';

export class DepositDto {
  @IsUUID()
  userId: string;

  @IsNumber()
  @Min(0.01)
  amount: number;

  @IsOptional()
  @IsString()
  description?: string;
}

export class TransferDto {
  @IsUUID()
  fromUserId: string;

  @IsUUID()
  toUserId: string;

  @IsNumber()
  @Min(0.01)
  amount: number;

  @IsOptional()
  @IsString()
  description?: string;
}

export class WalletResponseDto {
  id: string;
  userId: string;
  balance: number;
  createdAt: Date;
  updatedAt: Date;
}

export class TransactionResponseDto {
  id: string;
  fromWalletId?: string;
  toWalletId?: string;
  amount: number;
  type: TransactionType;
  description?: string;
  createdAt: Date;
}

