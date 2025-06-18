"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getSession = exports.getOrCreateSession = void 0;
const whatsapp_web_js_1 = require("whatsapp-web.js");
const qrcode_1 = __importDefault(require("qrcode"));
const sessions = new Map();
const getOrCreateSession = (sessionId) => {
    return new Promise((resolve, reject) => {
        if (sessions.has(sessionId)) {
            return reject(new Error('Session already active.'));
        }
        const client = new whatsapp_web_js_1.Client({
            authStrategy: new whatsapp_web_js_1.LocalAuth({ clientId: sessionId }),
            puppeteer: { headless: true, args: ['--no-sandbox'] },
        });
        client.on('qr', async (qr) => {
            const qrImage = await qrcode_1.default.toDataURL(qr);
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
exports.getOrCreateSession = getOrCreateSession;
const getSession = (sessionId) => {
    if (!sessions.has(sessionId)) {
        throw new Error('Session not found');
    }
    return sessions.get(sessionId);
};
exports.getSession = getSession;
