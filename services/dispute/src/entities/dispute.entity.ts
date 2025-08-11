import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';
import { DisputeStatus } from '../../../../shared/types/common.types';

@Entity('disputes')
export class Dispute {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid')
  contractId: string;

  @Column('uuid')
  raisedBy: string;

  @Column('text')
  description: string;

  @Column({
    type: 'enum',
    enum: DisputeStatus,
    default: DisputeStatus.OPEN,
  })
  status: DisputeStatus;

  @Column({ nullable: true })
  resolvedAt?: Date;

  @Column('text', { nullable: true })
  resolution?: string;

  @Column('uuid', { nullable: true })
  resolvedBy?: string;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  constructor(partial: Partial<Dispute>) {
    Object.assign(this, partial);
  }
}

