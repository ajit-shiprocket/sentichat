import { Client, LocalAuth } from 'whatsapp-web.js';
import qrcode from 'qrcode';

const sessions = new Map<string, Client>();

export const getOrCreateSession = (sessionId: string): Promise<string> => {
  return new Promise((resolve, reject) => {
    if (sessions.has(sessionId)) {
      return reject(new Error('Session already active.'));
    }

    const client = new Client({
      authStrategy: new LocalAuth({ clientId: sessionId }),
      puppeteer: { headless: true, args: ['--no-sandbox'] },
    });

    client.on('qr', async (qr) => {
      const qrImage = await qrcode.toDataURL(qr);
      resolve(qrImage);
    });

    client.on('ready', () => {
      console.log(`Client ${sessionId} is ready`);
    });

    client.on('message', async (msg) => {
      console.log(`[${sessionId}] Message received`, msg.body);
    });

    client.initialize();
    sessions.set(sessionId, client);
  });
};

export const getSession = (sessionId: string): Client => {
  if (!sessions.has(sessionId)) {
    throw new Error('Session not found');
  }
  return sessions.get(sessionId);
};