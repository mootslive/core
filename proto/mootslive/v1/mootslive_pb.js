// @generated by protoc-gen-es v0.3.0 with parameter "target=js+dts"
// @generated from file mootslive/v1/mootslive.proto (package mootslive.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from message mootslive.v1.GetStatusRequest
 */
export const GetStatusRequest = proto3.makeMessageType(
  "mootslive.v1.GetStatusRequest",
  [],
);

/**
 * @generated from message mootslive.v1.GetStatusResponse
 */
export const GetStatusResponse = proto3.makeMessageType(
  "mootslive.v1.GetStatusResponse",
  () => [
    { no: 1, name: "x_clacks_overhead", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message mootslive.v1.GetMeRequest
 */
export const GetMeRequest = proto3.makeMessageType(
  "mootslive.v1.GetMeRequest",
  [],
);

/**
 * @generated from message mootslive.v1.GetMeResponse
 */
export const GetMeResponse = proto3.makeMessageType(
  "mootslive.v1.GetMeResponse",
  () => [
    { no: 1, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "created_at", kind: "message", T: Timestamp },
  ],
);

/**
 * OAuth2State contains bits of state we need the client to hold during the
 * OAuth2 3-legged flow. These values aren't something that need to be kept
 * secret from the client.
 *
 * @generated from message mootslive.v1.OAuth2State
 */
export const OAuth2State = proto3.makeMessageType(
  "mootslive.v1.OAuth2State",
  () => [
    { no: 1, name: "state", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "pkce_code_verifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message mootslive.v1.BeginTwitterAuthRequest
 */
export const BeginTwitterAuthRequest = proto3.makeMessageType(
  "mootslive.v1.BeginTwitterAuthRequest",
  [],
);

/**
 * @generated from message mootslive.v1.BeginTwitterAuthResponse
 */
export const BeginTwitterAuthResponse = proto3.makeMessageType(
  "mootslive.v1.BeginTwitterAuthResponse",
  () => [
    { no: 1, name: "redirect_url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "state", kind: "message", T: OAuth2State },
  ],
);

/**
 * @generated from message mootslive.v1.FinishTwitterAuthRequest
 */
export const FinishTwitterAuthRequest = proto3.makeMessageType(
  "mootslive.v1.FinishTwitterAuthRequest",
  () => [
    { no: 1, name: "state", kind: "message", T: OAuth2State },
    { no: 2, name: "received_state", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "received_code", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message mootslive.v1.FinishTwitterAuthResponse
 */
export const FinishTwitterAuthResponse = proto3.makeMessageType(
  "mootslive.v1.FinishTwitterAuthResponse",
  () => [
    { no: 1, name: "me", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

