import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { ConfigProvider, theme } from 'antd';

const rootEl = document.getElementById('root');
if (rootEl) {
  const root = ReactDOM.createRoot(rootEl);
  root.render(
    <React.StrictMode>
      <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
        <App />
      </ConfigProvider>
    </React.StrictMode>
  )
}
