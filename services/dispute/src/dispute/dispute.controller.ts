import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Put,
  Delete,
  UseGuards,
  Request,
} from '@nestjs/common';
import { DisputeService } from './dispute.service';
import { CreateDisputeDto, ResolveDisputeDto, UpdateDisputeDto } from '../dto/dispute.dto';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@Controller('disputes')
@UseGuards(JwtAuthGuard)
export class DisputeController {
  constructor(private readonly disputeService: DisputeService) {}

  @Post()
  create(@Body() createDisputeDto: CreateDisputeDto) {
    return this.disputeService.create(createDisputeDto);
  }

  @Get(':id')
  findOne(@Param('id') id: string) {
    return this.disputeService.findOne(id);
  }

  @Get('contract/:contractId')
  findByContract(@Param('contractId') contractId: string) {
    return this.disputeService.findByContract(contractId);
  }

  @Get('user/:userId')
  findByUser(@Param('userId') userId: string) {
    return this.disputeService.findByUser(userId);
  }

  @Put(':id/resolve')
  resolve(
    @Param('id') id: string,
    @Body() resolveDisputeDto: ResolveDisputeDto,
    @Request() req: any,
  ) {
    return this.disputeService.resolve(id, resolveDisputeDto, req.user.userId);
  }

  @Put(':id')
  update(
    @Param('id') id: string,
    @Body() updateDisputeDto: UpdateDisputeDto,
    @Request() req: any,
  ) {
    return this.disputeService.update(id, updateDisputeDto, req.user.userId);
  }

  @Delete(':id')
  async delete(@Param('id') id: string, @Request() req: any) {
    await this.disputeService.delete(id, req.user.userId);
    return { message: 'Dispute deleted successfully' };
  }
}

