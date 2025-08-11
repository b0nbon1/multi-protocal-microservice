import {
  Injectable,
  NotFoundException,
  ForbiddenException,
  BadRequestException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Dispute } from '../entities/dispute.entity';
import { CreateDisputeDto, ResolveDisputeDto, UpdateDisputeDto } from '../dto/dispute.dto';
import { DisputeStatus } from '../../../../shared/types/common.types';

@Injectable()
export class DisputeService {
  constructor(
    @InjectRepository(Dispute)
    private disputeRepository: Repository<Dispute>,
  ) {}

  async create(createDisputeDto: CreateDisputeDto): Promise<Dispute> {
    // Check if dispute already exists for this contract
    const existingDispute = await this.disputeRepository.findOne({
      where: {
        contractId: createDisputeDto.contractId,
        status: DisputeStatus.OPEN,
      },
    });

    if (existingDispute) {
      throw new BadRequestException('An open dispute already exists for this contract');
    }

    const dispute = this.disputeRepository.create(createDisputeDto);
    return this.disputeRepository.save(dispute);
  }

  async findOne(id: string): Promise<Dispute> {
    const dispute = await this.disputeRepository.findOne({
      where: { id },
    });

    if (!dispute) {
      throw new NotFoundException('Dispute not found');
    }

    return dispute;
  }

  async findByContract(contractId: string): Promise<Dispute[]> {
    return this.disputeRepository.find({
      where: { contractId },
      order: { createdAt: 'DESC' },
    });
  }

  async findByUser(userId: string): Promise<Dispute[]> {
    return this.disputeRepository.find({
      where: { raisedBy: userId },
      order: { createdAt: 'DESC' },
    });
  }

  async resolve(
    id: string,
    resolveDisputeDto: ResolveDisputeDto,
    userId: string,
  ): Promise<Dispute> {
    const dispute = await this.findOne(id);

    if (dispute.status === DisputeStatus.RESOLVED) {
      throw new BadRequestException('Dispute is already resolved');
    }

    // For now, allow any authenticated user to resolve disputes
    // In a real system, you might want role-based access control
    dispute.status = DisputeStatus.RESOLVED;
    dispute.resolution = resolveDisputeDto.resolution;
    dispute.resolvedBy = resolveDisputeDto.resolvedBy;
    dispute.resolvedAt = new Date();

    return this.disputeRepository.save(dispute);
  }

  async update(
    id: string,
    updateDisputeDto: UpdateDisputeDto,
    userId: string,
  ): Promise<Dispute> {
    const dispute = await this.findOne(id);

    // Check if user is authorized to update this dispute
    if (dispute.raisedBy !== userId) {
      throw new ForbiddenException('You are not authorized to update this dispute');
    }

    if (dispute.status === DisputeStatus.RESOLVED) {
      throw new BadRequestException('Cannot update a resolved dispute');
    }

    Object.assign(dispute, updateDisputeDto);
    return this.disputeRepository.save(dispute);
  }

  async delete(id: string, userId: string): Promise<void> {
    const dispute = await this.findOne(id);

    // Check if user is authorized to delete this dispute
    if (dispute.raisedBy !== userId) {
      throw new ForbiddenException('You are not authorized to delete this dispute');
    }

    if (dispute.status === DisputeStatus.RESOLVED) {
      throw new BadRequestException('Cannot delete a resolved dispute');
    }

    await this.disputeRepository.remove(dispute);
  }
}

