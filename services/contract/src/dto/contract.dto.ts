import { IsString, IsUUID, IsNumber, IsEnum, IsOptional, Min } from 'class-validator';
import { ContractStatus } from '../../../../shared/types/common.types';

export class CreateContractDto {
  @IsUUID()
  sellerId: string;

  @IsUUID()
  buyerId: string;

  @IsString()
  title: string;

  @IsNumber()
  @Min(0.01)
  amount: number;

  @IsOptional()
  @IsEnum(ContractStatus)
  status?: ContractStatus;
}

export class UpdateContractDto {
  @IsOptional()
  @IsString()
  title?: string;

  @IsOptional()
  @IsNumber()
  @Min(0.01)
  amount?: number;

  @IsOptional()
  @IsEnum(ContractStatus)
  status?: ContractStatus;
}

