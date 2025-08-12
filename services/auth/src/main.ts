import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { Transport, MicroserviceOptions } from '@nestjs/microservices';
import { join } from 'path';
import { AppModule } from './app.module';

async function bootstrap() {
  // Create HTTP application
  const app = await NestFactory.create(AppModule);

  // Global validation pipe
  app.useGlobalPipes(new ValidationPipe({
    whitelist: true,
    forbidNonWhitelisted: true,
    transform: true,
  }));

  // Enable CORS
  app.enableCors({
    origin: process.env.CORS_ORIGIN || '*',
    credentials: true,
  });

  // Set global prefix for HTTP endpoints
  app.setGlobalPrefix('api/v1');

  const httpPort = process.env.PORT || 3001;
  const grpcPort = process.env.GRPC_PORT || 50051;

  // Create gRPC microservice
  const grpcApp = await NestFactory.createMicroservice<MicroserviceOptions>(AppModule, {
    transport: Transport.GRPC,
    options: {
      package: 'auth',
      protoPath: join(__dirname, '../../../proto/auth.proto'),
      url: `0.0.0.0:${grpcPort}`,
    },
  });

  // Start gRPC microservice
  await grpcApp.listen();
  console.log(`Auth gRPC Service is running on port ${grpcPort}`);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    console.log('SIGTERM received, shutting down gracefully');
    await app.close();
    await grpcApp.close();
    process.exit(0);
  });

  process.on('SIGINT', async () => {
    console.log('SIGINT received, shutting down gracefully');
    await app.close();
    await grpcApp.close();
    process.exit(0);
  });

  // Start HTTP server
  await app.listen(httpPort);
  console.log(`Auth HTTP Service is running on port ${httpPort}`);
}

bootstrap();

