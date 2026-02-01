import { NextResponse } from 'next/server';

export async function GET() {
  const healthCheck = {
    status: 'healthy',
    timestamp: new Date().toISOString(),
    service: 'frontend',
    version: process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0',
    uptime: process.uptime(),
    environment: process.env.NODE_ENV,
  };

  return NextResponse.json(healthCheck, { status: 200 });
}
