import { BeginTwitterAuthResponse } from "@mootslive/proto/mootslive/v1/mootslive_pb"
import React from "react"
import { createTransport, createUserServiceClient } from "../../modules/api"

const AuthTwitterPage = () => {
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
  return <div>Beginning twitter auth <a href={resp?.redirectUrl}>click me</a></div>
}

export default AuthTwitterPage