import mongoose from 'mongoose';

const chatSchema = new mongoose.Schema({
  sessionId: String,
  chatId: String,
  name: String,
  isGroup: Boolean
});

export default mongoose.model('Chat', chatSchema);