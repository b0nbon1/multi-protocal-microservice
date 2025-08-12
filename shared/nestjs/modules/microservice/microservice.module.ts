import { DynamicModule, Module } from '@nestjs/common';
import { MicroserviceModuleOptions, MicroserviceModuleAsyncOptions } from './grpc.interfaces';

@Module({})
export class MicroserviceModule {
  static forFeature(options: MicroserviceModuleOptions): DynamicModule {
    return {
      module: MicroserviceModule,
      providers: [
        {
          provide: 'MICROSERVICE_OPTIONS',
          useValue: options,
        },
      ],
      exports: ['MICROSERVICE_OPTIONS'],
    };
  }

  static forFeatureAsync(options: MicroserviceModuleAsyncOptions): DynamicModule {
    return {
      module: MicroserviceModule,
      imports: options.imports || [],
      providers: [
        {
          provide: 'MICROSERVICE_OPTIONS',
          useFactory: options.useFactory,
          inject: options.inject || [],
        },
      ],
      exports: ['MICROSERVICE_OPTIONS'],
    };
  }
}