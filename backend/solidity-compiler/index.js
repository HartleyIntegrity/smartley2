const express = require('express');
const solc = require('solc');
const app = express();
const port = 4000;

app.use(express.json());

app.post('/compile', (req, res) => {
  const { source } = req.body;

  if (!source) {
    res.status(400).json({ error: 'Missing Solidity source code' });
    return;
  }

  try {
    const input = {
      language: 'Solidity',
      sources: {
        'contract.sol': {
          content: source,
        },
      },
      settings: {
        outputSelection: {
          '*': {
            '*': ['*'],
          },
        },
      },
    };

    const output = JSON.parse(solc.compile(JSON.stringify(input)));

    if (output.errors && output.errors.length > 0) {
      res.status(400).json({ error: output.errors[0].formattedMessage });
      return;
    }

    res.json(output.contracts['contract.sol']);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.listen(port, () => {
  console.log(`Solidity compiler listening at http://localhost:${port}`);
});