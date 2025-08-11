import { IsString, IsUUID, IsEnum, IsOptional } from 'class-validator';
import { DisputeStatus } from '../../../../shared/types/common.types';

export class CreateDisputeDto {
  @IsUUID()
  contractId: string;

  @IsUUID()
  raisedBy: string;

  @IsString()
  description: string;
}

export class ResolveDisputeDto {
  @IsString()
  resolution: string;

  @IsUUID()
  resolvedBy: string;
}

export class UpdateDisputeDto {
  @IsOptional()
  @IsString()
  description?: string;

  @IsOptional()
  @IsEnum(DisputeStatus)
  status?: DisputeStatus;

  @IsOptional()
  @IsString()
  resolution?: string;
}

