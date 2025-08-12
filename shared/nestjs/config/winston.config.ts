// Import types dynamically to avoid build issues when packages aren't installed
export const winstonConfig = {
  level: 'info',
  format: {
    timestamp: true,
    colorize: true,
  },
  transports: ['console', 'file'],
  files: {
    error: 'logs/error.log',
    combined: 'logs/combined.log',
  },
};