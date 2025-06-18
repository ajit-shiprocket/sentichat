import express from 'express';
import dotenv from 'dotenv';
import mongoose from 'mongoose';
import { json } from 'body-parser';
import fileUpload from 'express-fileupload';
import swaggerUi from 'swagger-ui-express';
import swaggerSpec from './swagger/swagger';
import messageRoutes from './routes/messageRoutes';
import authRoutes from './routes/authRoutes';
import chatRoutes from './routes/chatRoutes';

dotenv.config();

const app = express();
app.use(json());
app.use(fileUpload());

app.use('/api/messages', messageRoutes);
app.use('/api/auth', authRoutes);
app.use('/api/chats', chatRoutes);

app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec));

mongoose.connect(process.env.MONGO_URI || '', {}).then(() => {
  console.log('Connected to MongoDB');
});

export default app;