// Automatically instruments using the default OpenCensus plugins.
const instrument = require('./instrument');
instrument.start({
  serviceName: 'service-b',
  agentPort: 55678,
});

const express = require('express');
const http = require('http');

const port = process.env.PORT || 8081
const app = express();

app.get('/sleep/:timeout', async (request, response) => {
  const timeout = request.params.timeout
  await new Promise((resolve, reject) => setTimeout(resolve, timeout))
  response.status(200);
  response.send(`okay from service-b with timeout ${timeout}`);
});

http
  .createServer(app)
  .listen(port, () => {
    console.log(`Your app is listening on http://localhost:${port}`);
  });
