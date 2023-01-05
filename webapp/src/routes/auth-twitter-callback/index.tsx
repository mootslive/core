import { FinishTwitterAuthResponse, OAuth2State } from "@mootslive/proto/mootslive/v1/mootslive_pb"
import React from "react"
import { useSearchParams } from "react-router-dom"
import { createTransport, createUserServiceClient } from "../../modules/api"

const AuthTwitterCallbackPage = () => {
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
  
export default AuthTwitterCallbackPage