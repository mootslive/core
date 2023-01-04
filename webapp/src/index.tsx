import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import reportWebVitals from './reportWebVitals';
import {
  createBrowserRouter,
  RouterProvider,
  useSearchParams
} from "react-router-dom";
import {
  createConnectTransport,
  createPromiseClient,
  Transport
} from "@bufbuild/connect-web";
import {
  UserService,
} from "@mootslive/proto/mootslive/v1/mootslive_connectweb"
import { BeginTwitterAuthResponse, FinishTwitterAuthResponse, OAuth2State } from '@mootslive/proto/mootslive/v1/mootslive_pb';

const createTransport = () => {
  return createConnectTransport({
    baseUrl: "http://localhost:9000"
  })
}

const createUserServiceClient = (t: Transport) => {
  return createPromiseClient(UserService, t)
}

const BeginTwitterAuthPage = () => {
  const client = createUserServiceClient(createTransport())

  const [resp, setResp] = React.useState<BeginTwitterAuthResponse>()
  React.useEffect(() => {
    client.beginTwitterAuth({}).then((resp) => {
      setResp(resp)
      if (!resp || !resp.state) {
        return
      }
      localStorage.setItem("twitter_auth_state", JSON.stringify(resp.state))
    })
  }, [])
  return <div>Beginning twitter auth {resp?.redirectUrl}</div>
}

const FinishTwitterAuthPage = () => {
  const client = createUserServiceClient(createTransport())

  const [queryParams] = useSearchParams()

  const state = queryParams.get("state")
  if (!state) {
    throw Error("missing state")
  }

  const code = queryParams.get("code")
  if (!code) {
    throw Error("missing code")
  }

  const [storedState] = React.useState(() => {
    // getting stored value
    const saved = localStorage.getItem("twitter_auth_state");
    if (!saved) {
      throw new Error("no localstorage state")
    }
    const initialValue = JSON.parse(saved) as OAuth2State;
    return initialValue;
  });
  

 const [resp, setResp] = React.useState<FinishTwitterAuthResponse>()
  React.useEffect(() => {
    client.finishTwitterAuth({
      receivedState: state,
      receivedCode: code,
      state: storedState,
    }).then((resp) => {
      setResp(resp)
    })
  }, [code, state, storedState])

  return <div>Finishing twitter auth <br/><br/> {resp ? resp.me: <strong>loading...</strong>}</div>
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <div>hello world</div>
  },
  {
    path: "/auth/twitter",
    element: <BeginTwitterAuthPage/>
  },
  {
    path: "/auth/twitter/callback",
    element: <FinishTwitterAuthPage/>
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
