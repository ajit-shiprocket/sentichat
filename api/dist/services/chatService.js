"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getChats = void 0;
const sessionService_1 = require("./sessionService");
const getChats = async (sessionId) => {
    const client = (0, sessionService_1.getSession)(sessionId);
    const chats = await client.getChats();
    return chats.map(chat => ({
        chatId: chat.id._serialized,
        name: chat.name,
        isGroup: chat.isGroup
    }));
};
exports.getChats = getChats;
