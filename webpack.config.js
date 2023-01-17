const path = require('path');

module.exports = (env, argv) => {
  const config = {
    entry: './src/ts/main.ts',
    module: {
      rules: [
        {
          test: /\.tsx?$/,
          use: 'ts-loader',
          exclude: /node_modules/,
        },
      ],
    },
    resolve: {
      extensions: ['.tsx', '.ts', '.js'],
    },
    output: {
      filename: 'app.bundle.js',
      path: path.resolve(__dirname, 'src/js'),
    },
  };

  if (argv.mode !== 'production') {
    config.devtool = 'inline-source-map';
  }

  return config;
};
