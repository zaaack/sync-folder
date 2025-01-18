# README

## About

A simple and low memory(10mb) tray application for back up of your folders.

![](./docs/QQ20250118-091740.png)

## Live Development

To run in live development mode, run `wails dev` in the project directory, and exit tray app several times. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`, and exit tray app when stuck to continue.
