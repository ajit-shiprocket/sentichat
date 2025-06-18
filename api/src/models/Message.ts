import mongoose from 'mongoose';

const messageSchema = new mongoose.Schema({
  sessionId: String,
  chatId: String,
  body: String,
  type: String,
  media: String,
  timestamp: Date
});

export default mongoose.model('Message', messageSchema);