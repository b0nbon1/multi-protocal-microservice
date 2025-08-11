import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  ManyToOne,
  JoinColumn,
} from 'typeorm';
import { TransactionType } from '../../../../shared/types/common.types';
import { Wallet } from './wallet.entity';

@Entity('transactions')
export class Transaction {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid', { nullable: true })
  fromWalletId?: string;

  @Column('uuid', { nullable: true })
  toWalletId?: string;

  @Column('decimal', { precision: 10, scale: 2 })
  amount: number;

  @Column({
    type: 'enum',
    enum: TransactionType,
  })
  type: TransactionType;

  @Column({ nullable: true })
  description?: string;

  @CreateDateColumn()
  createdAt: Date;

  @ManyToOne(() => Wallet, wallet => wallet.outgoingTransactions)
  @JoinColumn({ name: 'fromWalletId' })
  fromWallet?: Wallet;

  @ManyToOne(() => Wallet, wallet => wallet.incomingTransactions)
  @JoinColumn({ name: 'toWalletId' })
  toWallet?: Wallet;

  constructor(partial: Partial<Transaction>) {
    Object.assign(this, partial);
  }
}

