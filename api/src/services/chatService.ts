import { getSession } from './sessionService';

export const getChats = async (sessionId: string) => {
  const client = getSession(sessionId);
  const chats = await client.getChats();
  return chats.map(chat => ({
    chatId: chat.id._serialized,
    name: chat.name,
    isGroup: chat.isGroup
  }));
};