import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import reportWebVitals from './reportWebVitals';
import {
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";
import AuthTwitterPage from './routes/auth-twitter';
import AuthTwitterCallbackPage from './routes/auth-twitter-callback';


const router = createBrowserRouter([
  {
    path: "/",
    element: <div>hello world</div>
  },
  {
    path: "/auth/twitter",
    element: <AuthTwitterPage/>
  },
  {
    path: "/auth/twitter/callback",
    element: <AuthTwitterCallbackPage/>
  }
])

ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
