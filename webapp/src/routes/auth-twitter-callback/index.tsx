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

  const savedStateJSON = localStorage.getItem("twitter_auth_state");
  if (!savedStateJSON) {
    throw new Error("no localstorage state")
  }
  const storedState = JSON.parse(savedStateJSON) as OAuth2State;


  const [resp, setResp] = React.useState<FinishTwitterAuthResponse>()
  // We use a Ref here to ensure this only runs once even in a remount.
  // This is because running this request twice invalidates the authorization
  // code.
  const authAttempted = React.useRef(false)
  React.useEffect(() => {
    if (!authAttempted.current) {
      authAttempted.current = true
      client.finishTwitterAuth({
        receivedState: state,
        receivedCode: code,
        state: storedState,
      }).then((resp) => {
        setResp(resp)
      })
    }
  }, [state, code, client, storedState])

  return <div>Finishing twitter auth <br/><br/> moots id token {resp ? resp.idToken: <strong>loading...</strong>}</div>
}
  
export default AuthTwitterCallbackPage