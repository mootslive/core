import {
  createConnectTransport,
  createPromiseClient,
  Transport
} from "@bufbuild/connect-web";
import {
  UserService,
} from "@mootslive/proto/mootslive/v1/mootslive_connectweb"

export const createTransport = () => {
    return createConnectTransport({
      baseUrl: "http://localhost:9000"
    })
  }
  
export const createUserServiceClient = (t: Transport) => {
  return createPromiseClient(UserService, t)
}
  