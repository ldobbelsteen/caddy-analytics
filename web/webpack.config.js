const path = require("path")
const HtmlWebpackPlugin = require("html-webpack-plugin")
const { CleanWebpackPlugin } = require("clean-webpack-plugin")
const ESLintPlugin = require("eslint-webpack-plugin")

module.exports = {
	entry: "./src/main.js",
	plugins: [
		new ESLintPlugin(),
		new CleanWebpackPlugin(),
		new HtmlWebpackPlugin({
			template: "./src/index.html",
			scriptLoading: "defer"
		})
	],
	output: {
		path: path.resolve(__dirname, "dist"),
		filename: "[contenthash].js",
	},
	module: {
		rules: [
			{
				test: /\.css$/i,
				use: ["style-loader", "css-loader"]
			},
			{
				test: /\.sass$/i,
				use: ["style-loader", "css-loader", "sass-loader"]
			},
			{
				test: /\.(woff|woff2)$/i,
				type: "asset/resource"
			}
		]
	}
}
