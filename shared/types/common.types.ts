export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginationQuery {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: 'ASC' | 'DESC';
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export enum ServiceStatus {
  HEALTHY = 'healthy',
  UNHEALTHY = 'unhealthy',
  DEGRADED = 'degraded'
}

export interface HealthCheck {
  service: string;
  status: ServiceStatus;
  timestamp: Date;
  version?: string;
  uptime?: number;
  dependencies?: Record<string, ServiceStatus>;
}

export interface JwtPayload {
  sub: string;
  email: string;
  iat?: number;
  exp?: number;
}

export interface User {
  id: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

export enum ContractStatus {
  DRAFT = 'DRAFT',
  ACTIVE = 'ACTIVE',
  COMPLETED = 'COMPLETED'
}

export interface Contract {
  id: string;
  sellerId: string;
  buyerId: string;
  title: string;
  amount: number;
  status: ContractStatus;
  createdAt: Date;
  updatedAt: Date;
}

export enum TransactionType {
  DEPOSIT = 'DEPOSIT',
  WITHDRAWAL = 'WITHDRAWAL',
  TRANSFER = 'TRANSFER',
  PAYMENT = 'PAYMENT',
  REFUND = 'REFUND'
}

export interface Wallet {
  id: string;
  userId: string;
  balance: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface Transaction {
  id: string;
  fromWalletId?: string;
  toWalletId?: string;
  amount: number;
  type: TransactionType;
  description?: string;
  createdAt: Date;
}

export enum DisputeStatus {
  OPEN = 'OPEN',
  RESOLVED = 'RESOLVED'
}

export interface Dispute {
  id: string;
  contractId: string;
  raisedBy: string;
  description: string;
  status: DisputeStatus;
  resolvedAt?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface AuditLog {
  id: string;
  userId?: string;
  action: string;
  metadata?: Record<string, any>;
  timestamp: Date;
}

export interface NotificationMessage {
  id: string;
  userId?: string;
  type: string;
  title: string;
  message: string;
  data?: Record<string, any>;
  timestamp: Date;
}

