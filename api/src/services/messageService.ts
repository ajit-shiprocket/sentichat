import { MessageMedia } from 'whatsapp-web.js';
import { getSession } from './sessionService';
import Message from '../models/Message';
import fs from 'fs';

export const sendTextMessage = async ({ sessionId, to, message }) => {
  const client = getSession(sessionId);
  const sentMsg = await client.sendMessage(to, message);
  return await Message.create({
    sessionId,
    chatId: to,
    body: message,
    type: 'text',
    timestamp: new Date(),
  });
};

export const sendMediaMessage = async (req) => {
  const { sessionId, to } = req.body;
  const file = req.files.media;
  const client = getSession(sessionId);

  const media = new MessageMedia(file.mimetype, file.data.toString('base64'), file.name);
  await client.sendMessage(to, media);

  return await Message.create({
    sessionId,
    chatId: to,
    type: 'media',
    media: file.name,
    timestamp: new Date(),
  });
};