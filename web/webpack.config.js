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
		filename: "[name].[contenthash].js",
	},
	module: {
		rules: [
			{
				test: /\.js$/,
				exclude: /node_modules/,
				use: {
					loader: "babel-loader",
					options: {
						presets: [
							[
								"@babel/preset-env", {
									"targets": "> 0.25%, not dead",
									"useBuiltIns": "usage",
									"corejs": 3
								}
							]
						]
					}
				}
			},
			{
				test: /\.css$/i,
				use: ["style-loader", "css-loader"]
			},
			{
				test: /\.sass$/i,
				use: ["style-loader", "css-loader", "sass-loader"]
			},
			{
				test: /\.(woff2|woff|eot|ttf|otf)$/i,
				type: "asset/resource"
			}
		]
	}
}
