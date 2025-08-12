import { applyDecorators } from '@nestjs/common';
import { GrpcMethod } from '@nestjs/microservices';

export function GrpcService(service: string) {
  return applyDecorators();
}

export function GrpcRpc(method: string) {
  return (target: any, propertyKey: string, descriptor: PropertyDescriptor) => {
    return GrpcMethod(method)(target, propertyKey, descriptor);
  };
}