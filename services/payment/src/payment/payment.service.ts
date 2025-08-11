import {
  Injectable,
  NotFoundException,
  BadRequestException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, DataSource } from 'typeorm';
import { Wallet } from '../entities/wallet.entity';
import { Transaction } from '../entities/transaction.entity';
import { DepositDto, TransferDto } from '../dto/payment.dto';
import { TransactionType } from '../../../../shared/types/common.types';

@Injectable()
export class PaymentService {
  constructor(
    @InjectRepository(Wallet)
    private walletRepository: Repository<Wallet>,
    @InjectRepository(Transaction)
    private transactionRepository: Repository<Transaction>,
    private dataSource: DataSource,
  ) {}

  async getOrCreateWallet(userId: string): Promise<Wallet> {
    let wallet = await this.walletRepository.findOne({
      where: { userId },
    });

    if (!wallet) {
      wallet = this.walletRepository.create({
        userId,
        balance: 0,
      });
      wallet = await this.walletRepository.save(wallet);
    }

    return wallet;
  }

  async getWalletBalance(userId: string): Promise<Wallet> {
    return this.getOrCreateWallet(userId);
  }

  async deposit(depositDto: DepositDto): Promise<Transaction> {
    const { userId, amount, description } = depositDto;

    // Use transaction to ensure data consistency
    return this.dataSource.transaction(async manager => {
      const wallet = await manager.findOne(Wallet, {
        where: { userId },
      }) || await manager.save(Wallet, {
        userId,
        balance: 0,
      });

      // Update wallet balance
      wallet.balance = Number(wallet.balance) + Number(amount);
      await manager.save(wallet);

      // Create transaction record
      const transaction = manager.create(Transaction, {
        toWalletId: wallet.id,
        amount,
        type: TransactionType.DEPOSIT,
        description,
      });

      return manager.save(transaction);
    });
  }

  async transfer(transferDto: TransferDto): Promise<Transaction> {
    const { fromUserId, toUserId, amount, description } = transferDto;

    if (fromUserId === toUserId) {
      throw new BadRequestException('Cannot transfer to the same wallet');
    }

    // Use transaction to ensure data consistency
    return this.dataSource.transaction(async manager => {
      // Get or create sender wallet
      let fromWallet = await manager.findOne(Wallet, {
        where: { userId: fromUserId },
      });

      if (!fromWallet) {
        throw new NotFoundException('Sender wallet not found');
      }

      // Check if sender has sufficient balance
      if (Number(fromWallet.balance) < Number(amount)) {
        throw new BadRequestException('Insufficient balance');
      }

      // Get or create receiver wallet
      let toWallet = await manager.findOne(Wallet, {
        where: { userId: toUserId },
      }) || await manager.save(Wallet, {
        userId: toUserId,
        balance: 0,
      });

      // Update balances
      fromWallet.balance = Number(fromWallet.balance) - Number(amount);
      toWallet.balance = Number(toWallet.balance) + Number(amount);

      await manager.save([fromWallet, toWallet]);

      // Create transaction record
      const transaction = manager.create(Transaction, {
        fromWalletId: fromWallet.id,
        toWalletId: toWallet.id,
        amount,
        type: TransactionType.TRANSFER,
        description,
      });

      return manager.save(transaction);
    });
  }

  async getTransactionHistory(userId: string): Promise<Transaction[]> {
    const wallet = await this.getOrCreateWallet(userId);

    return this.transactionRepository.find({
      where: [
        { fromWalletId: wallet.id },
        { toWalletId: wallet.id },
      ],
      order: { createdAt: 'DESC' },
    });
  }
}

