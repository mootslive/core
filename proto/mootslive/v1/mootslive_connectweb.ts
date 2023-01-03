// @generated by protoc-gen-connect-web v0.3.3 with parameter "target=ts"
// @generated from file mootslive/v1/mootslive.proto (package mootslive.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { BeginTwitterAuthRequest, BeginTwitterAuthResponse, FinishTwitterAuthRequest, FinishTwitterAuthResponse, GetMeRequest, GetMeResponse, GetStatusRequest, GetStatusResponse } from "./mootslive_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service mootslive.v1.AdminService
 */
export const AdminService = {
  typeName: "mootslive.v1.AdminService",
  methods: {
    /**
     * @generated from rpc mootslive.v1.AdminService.GetStatus
     */
    getStatus: {
      name: "GetStatus",
      I: GetStatusRequest,
      O: GetStatusResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * @generated from service mootslive.v1.UserService
 */
export const UserService = {
  typeName: "mootslive.v1.UserService",
  methods: {
    /**
     * @generated from rpc mootslive.v1.UserService.GetMe
     */
    getMe: {
      name: "GetMe",
      I: GetMeRequest,
      O: GetMeResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mootslive.v1.UserService.BeginTwitterAuth
     */
    beginTwitterAuth: {
      name: "BeginTwitterAuth",
      I: BeginTwitterAuthRequest,
      O: BeginTwitterAuthResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mootslive.v1.UserService.FinishTwitterAuth
     */
    finishTwitterAuth: {
      name: "FinishTwitterAuth",
      I: FinishTwitterAuthRequest,
      O: FinishTwitterAuthResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

