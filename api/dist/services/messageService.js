"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.sendMediaMessage = exports.sendTextMessage = void 0;
const whatsapp_web_js_1 = require("whatsapp-web.js");
const sessionService_1 = require("./sessionService");
const Message_1 = __importDefault(require("../models/Message"));
const sendTextMessage = async ({ sessionId, to, message }) => {
    const client = (0, sessionService_1.getSession)(sessionId);
    const sentMsg = await client.sendMessage(to, message);
    return await Message_1.default.create({
        sessionId,
        chatId: to,
        body: message,
        type: 'text',
        timestamp: new Date(),
    });
};
exports.sendTextMessage = sendTextMessage;
const sendMediaMessage = async (req) => {
    const { sessionId, to } = req.body;
    const file = req.files.media;
    const client = (0, sessionService_1.getSession)(sessionId);
    const media = new whatsapp_web_js_1.MessageMedia(file.mimetype, file.data.toString('base64'), file.name);
    await client.sendMessage(to, media);
    return await Message_1.default.create({
        sessionId,
        chatId: to,
        type: 'media',
        media: file.name,
        timestamp: new Date(),
    });
};
exports.sendMediaMessage = sendMediaMessage;
