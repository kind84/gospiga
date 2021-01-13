const path = require('path');
const webpack = require('webpack');

module.exports = {
  entry: './src/recipes.js',
  output: {
    filename: 'recipes.js',
    path: path.resolve(__dirname, 'dist'),
  },
  target: 'webworker',
  plugins: [
    new webpack.IgnorePlugin(/^hiredis$/),
    new webpack.IgnorePlugin(/^net$/),
    new webpack.IgnorePlugin(/^tls$/)
  ],
};
