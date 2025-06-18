import { Request, Response } from 'express';
import * as chatService from '../services/chatService';

export const getChats = async (req: Request, res: Response) => {
  try {
    const sessionId = req.params.sessionId;
    const chats = await chatService.getChats(sessionId);
    res.json(chats);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
};