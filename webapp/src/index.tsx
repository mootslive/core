import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import reportWebVitals from './reportWebVitals';
import {
  createBrowserRouter,
  RouterProvider
} from "react-router-dom"

const BeginTwitterAuthPage = () => {
  return <div>Beginning twitter auth</div>
}

const FinishTwitterAuthPage = () => {
  return <div>Finishing twitter auth</div>
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <div>hello world</div>
  },
])

const root = ReactDOM.createRoot(
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
