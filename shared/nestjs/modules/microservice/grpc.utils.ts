import { RpcException } from '@nestjs/microservices';
import { status } from '@grpc/grpc-js';
import { GrpcAuthResponse, GrpcUser } from './grpc.interfaces';

export class GrpcUtils {
  static createSuccessResponse(data?: any, message?: string): any {
    return {
      success: true,
      message,
      ...data,
    };
  }

  static createErrorResponse(message: string, code?: status): any {
    throw new RpcException({
      code: code || status.INTERNAL,
      message,
    });
  }

  static createAuthResponse(
    success: boolean,
    user?: any,
    accessToken?: string,
    message?: string,
  ): GrpcAuthResponse {
    const response: GrpcAuthResponse = {
      success,
      message,
    };

    if (user) {
      response.user = {
        id: user.id,
        email: user.email,
        firstName: user.firstName,
        lastName: user.lastName,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
      };
    }

    if (accessToken) {
      response.accessToken = accessToken;
    }

    return response;
  }

  static formatUser(user: any): GrpcUser {
    return {
      id: user.id,
      email: user.email,
      firstName: user.firstName,
      lastName: user.lastName,
      createdAt: user.createdAt,
      updatedAt: user.updatedAt,
    };
  }
}