const swaggerSpec = {
  openapi: '3.0.0',
  info: {
    title: 'WhatsApp API',
    version: '1.0.0',
    description: 'Multi-session WhatsApp API with media and QR login',
  },
  servers: [
    {
      url: 'http://localhost:3000',
    },
  ],
};

export default swaggerSpec;