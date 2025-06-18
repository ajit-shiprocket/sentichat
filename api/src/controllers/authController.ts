import { Request, Response } from 'express';
import * as sessionService from '../services/sessionService';

export const generateQR = async (req: Request, res: Response) => {
  try {
    const sessionId = req.params.sessionId;
    const qr = await sessionService.getOrCreateSession(sessionId);
    res.json({ sessionId, qr });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
};