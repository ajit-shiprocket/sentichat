"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const dotenv_1 = __importDefault(require("dotenv"));
const mongoose_1 = __importDefault(require("mongoose"));
const body_parser_1 = require("body-parser");
const express_fileupload_1 = __importDefault(require("express-fileupload"));
const swagger_ui_express_1 = __importDefault(require("swagger-ui-express"));
const swagger_1 = __importDefault(require("./swagger/swagger"));
const messageRoutes_1 = __importDefault(require("./routes/messageRoutes"));
const authRoutes_1 = __importDefault(require("./routes/authRoutes"));
const chatRoutes_1 = __importDefault(require("./routes/chatRoutes"));
dotenv_1.default.config();
const app = (0, express_1.default)();
app.use((0, body_parser_1.json)());
app.use((0, express_fileupload_1.default)());
app.use('/api/messages', messageRoutes_1.default);
app.use('/api/auth', authRoutes_1.default);
app.use('/api/chats', chatRoutes_1.default);
app.use('/api-docs', swagger_ui_express_1.default.serve, swagger_ui_express_1.default.setup(swagger_1.default));
mongoose_1.default.connect(process.env.MONGO_URI || '', {}).then(() => {
    console.log('Connected to MongoDB');
});
exports.default = app;
