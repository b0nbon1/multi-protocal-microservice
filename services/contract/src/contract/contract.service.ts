import {
  Injectable,
  NotFoundException,
  ForbiddenException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Contract } from '../entities/contract.entity';
import { CreateContractDto, UpdateContractDto } from '../dto/contract.dto';

@Injectable()
export class ContractService {
  constructor(
    @InjectRepository(Contract)
    private contractRepository: Repository<Contract>,
  ) {}

  async create(createContractDto: CreateContractDto): Promise<Contract> {
    const contract = this.contractRepository.create(createContractDto);
    return this.contractRepository.save(contract);
  }

  async findOne(id: string): Promise<Contract> {
    const contract = await this.contractRepository.findOne({
      where: { id },
    });

    if (!contract) {
      throw new NotFoundException('Contract not found');
    }

    return contract;
  }

  async findByUser(userId: string): Promise<Contract[]> {
    return this.contractRepository.find({
      where: [
        { sellerId: userId },
        { buyerId: userId },
      ],
      order: { createdAt: 'DESC' },
    });
  }

  async update(
    id: string,
    updateContractDto: UpdateContractDto,
    userId: string,
  ): Promise<Contract> {
    const contract = await this.findOne(id);

    // Check if user is authorized to update this contract
    if (contract.sellerId !== userId && contract.buyerId !== userId) {
      throw new ForbiddenException('You are not authorized to update this contract');
    }

    Object.assign(contract, updateContractDto);
    return this.contractRepository.save(contract);
  }

  async delete(id: string, userId: string): Promise<void> {
    const contract = await this.findOne(id);

    // Check if user is authorized to delete this contract
    if (contract.sellerId !== userId && contract.buyerId !== userId) {
      throw new ForbiddenException('You are not authorized to delete this contract');
    }

    await this.contractRepository.remove(contract);
  }
}

