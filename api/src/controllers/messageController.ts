import { Request, Response } from 'express';
import * as messageService from '../services/messageService';

export const sendMessage = async (req: Request, res: Response) => {
  try {
    const response = await messageService.sendTextMessage(req.body);
    res.json(response);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
};

export const sendMediaMessage = async (req: Request, res: Response) => {
  try {
    const response = await messageService.sendMediaMessage(req);
    res.json(response);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
};