import { ModuleMetadata } from '@nestjs/common';

export interface MicroserviceModuleOptions {
  name: string;
  protoPath: string;
  package: string;
  url: string;
  port: number;
}

export interface MicroserviceModuleAsyncOptions
  extends Pick<ModuleMetadata, 'imports'> {
  useFactory: (
    ...args: any[]
  ) => Promise<MicroserviceModuleOptions> | MicroserviceModuleOptions;
  inject?: any[];
}

export interface GrpcUser {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface GrpcAuthResponse {
  success: boolean;
  message?: string;
  accessToken?: string;
  user?: GrpcUser;
}

export interface GrpcValidationResponse {
  isValid: boolean;
  user?: GrpcUser;
  message?: string;
}