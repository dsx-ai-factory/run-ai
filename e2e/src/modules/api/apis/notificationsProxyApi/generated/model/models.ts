import localVarRequest from 'request';

export * from './connectSlackWorkspace200Response';
export * from './createSlackAppRequestBody';
export * from './createSlackAppResponse';
export * from './detailedNotificationState';
export * from './detailedNotificationStateSlack';
export * from './detailedSlackNotificationsState';
export * from './emailAuthType';
export * from './emailChannel';
export * from './emailChannelConfig';
export * from './emailChannelConfigForPatch';
export * from './emailChannelForPatch';
export * from './emailNotificationChannelConfig';
export * from './modelError';
export * from './notificationChannel';
export * from './notificationChannelConfig';
export * from './notificationChannelForPatch';
export * from './notificationChannelForPatchConfig';
export * from './notificationChannelForValidate';
export * from './notificationChannelMeta';
export * from './notificationChannelResponse';
export * from './notificationChannelResponseAllOf';
export * from './notificationChannelType';
export * from './notificationState';
export * from './notificationStateStatus';
export * from './notificationType';
export * from './slackDetails';
export * from './slackDetailsSettings';
export * from './slackNotificationChannel';
export * from './slackNotificationChannelConfig';
export * from './slackNotificationChannelPatchRequest';
export * from './slackNotificationChannelPatchRequestConfig';
export * from './subscription';
export * from './subscriptionAllOf';
export * from './subscriptionEvent';
export * from './subscriptionForPut';
export * from './validateSinkResponse';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { ConnectSlackWorkspace200Response } from './connectSlackWorkspace200Response';
import { CreateSlackAppRequestBody } from './createSlackAppRequestBody';
import { CreateSlackAppResponse } from './createSlackAppResponse';
import { DetailedNotificationState } from './detailedNotificationState';
import { DetailedNotificationStateSlack } from './detailedNotificationStateSlack';
import { DetailedSlackNotificationsState } from './detailedSlackNotificationsState';
import { EmailAuthType } from './emailAuthType';
import { EmailChannel } from './emailChannel';
import { EmailChannelConfig } from './emailChannelConfig';
import { EmailChannelConfigForPatch } from './emailChannelConfigForPatch';
import { EmailChannelForPatch } from './emailChannelForPatch';
import { EmailNotificationChannelConfig } from './emailNotificationChannelConfig';
import { ModelError } from './modelError';
import { NotificationChannel } from './notificationChannel';
import { NotificationChannelConfig } from './notificationChannelConfig';
import { NotificationChannelForPatch } from './notificationChannelForPatch';
import { NotificationChannelForPatchConfig } from './notificationChannelForPatchConfig';
import { NotificationChannelForValidate } from './notificationChannelForValidate';
import { NotificationChannelMeta } from './notificationChannelMeta';
import { NotificationChannelResponse } from './notificationChannelResponse';
import { NotificationChannelResponseAllOf } from './notificationChannelResponseAllOf';
import { NotificationChannelType } from './notificationChannelType';
import { NotificationState } from './notificationState';
import { NotificationStateStatus } from './notificationStateStatus';
import { NotificationType } from './notificationType';
import { SlackDetails } from './slackDetails';
import { SlackDetailsSettings } from './slackDetailsSettings';
import { SlackNotificationChannel } from './slackNotificationChannel';
import { SlackNotificationChannelConfig } from './slackNotificationChannelConfig';
import { SlackNotificationChannelPatchRequest } from './slackNotificationChannelPatchRequest';
import { SlackNotificationChannelPatchRequestConfig } from './slackNotificationChannelPatchRequestConfig';
import { Subscription } from './subscription';
import { SubscriptionAllOf } from './subscriptionAllOf';
import { SubscriptionEvent } from './subscriptionEvent';
import { SubscriptionForPut } from './subscriptionForPut';
import { ValidateSinkResponse } from './validateSinkResponse';

/* tslint:disable:no-unused-variable */
let primitives = [
                    "string",
                    "boolean",
                    "double",
                    "integer",
                    "long",
                    "float",
                    "number",
                    "any"
                 ];

let enumsMap: {[index: string]: any} = {
        "EmailAuthType": EmailAuthType,
        "NotificationChannelMeta.TypeEnum": NotificationChannelMeta.TypeEnum,
        "NotificationChannelType": NotificationChannelType,
}

let typeMap: {[index: string]: any} = {
    "ConnectSlackWorkspace200Response": ConnectSlackWorkspace200Response,
    "CreateSlackAppRequestBody": CreateSlackAppRequestBody,
    "CreateSlackAppResponse": CreateSlackAppResponse,
    "DetailedNotificationState": DetailedNotificationState,
    "DetailedNotificationStateSlack": DetailedNotificationStateSlack,
    "DetailedSlackNotificationsState": DetailedSlackNotificationsState,
    "EmailChannel": EmailChannel,
    "EmailChannelConfig": EmailChannelConfig,
    "EmailChannelConfigForPatch": EmailChannelConfigForPatch,
    "EmailChannelForPatch": EmailChannelForPatch,
    "EmailNotificationChannelConfig": EmailNotificationChannelConfig,
    "ModelError": ModelError,
    "NotificationChannel": NotificationChannel,
    "NotificationChannelConfig": NotificationChannelConfig,
    "NotificationChannelForPatch": NotificationChannelForPatch,
    "NotificationChannelForPatchConfig": NotificationChannelForPatchConfig,
    "NotificationChannelForValidate": NotificationChannelForValidate,
    "NotificationChannelMeta": NotificationChannelMeta,
    "NotificationChannelResponse": NotificationChannelResponse,
    "NotificationChannelResponseAllOf": NotificationChannelResponseAllOf,
    "NotificationState": NotificationState,
    "NotificationStateStatus": NotificationStateStatus,
    "NotificationType": NotificationType,
    "SlackDetails": SlackDetails,
    "SlackDetailsSettings": SlackDetailsSettings,
    "SlackNotificationChannel": SlackNotificationChannel,
    "SlackNotificationChannelConfig": SlackNotificationChannelConfig,
    "SlackNotificationChannelPatchRequest": SlackNotificationChannelPatchRequest,
    "SlackNotificationChannelPatchRequestConfig": SlackNotificationChannelPatchRequestConfig,
    "Subscription": Subscription,
    "SubscriptionAllOf": SubscriptionAllOf,
    "SubscriptionEvent": SubscriptionEvent,
    "SubscriptionForPut": SubscriptionForPut,
    "ValidateSinkResponse": ValidateSinkResponse,
}

export class ObjectSerializer {
    public static findCorrectType(data: any, expectedType: string) {
        if (data == undefined) {
            return expectedType;
        } else if (primitives.indexOf(expectedType.toLowerCase()) !== -1) {
            return expectedType;
        } else if (expectedType === "Date") {
            return expectedType;
        } else {
            if (enumsMap[expectedType]) {
                return expectedType;
            }

            if (!typeMap[expectedType]) {
                return expectedType; // w/e we don't know the type
            }

            // Check the discriminator
            let discriminatorProperty = typeMap[expectedType].discriminator;
            if (discriminatorProperty == null) {
                return expectedType; // the type does not have a discriminator. use it.
            } else {
                if (data[discriminatorProperty]) {
                    var discriminatorType = data[discriminatorProperty];
                    if(typeMap[discriminatorType]){
                        return discriminatorType; // use the type given in the discriminator
                    } else {
                        return expectedType; // discriminator did not map to a type
                    }
                } else {
                    return expectedType; // discriminator was not present (or an empty string)
                }
            }
        }
    }

    public static serialize(data: any, type: string) {
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.serialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return data.toISOString();
        } else {
            if (enumsMap[type]) {
                return data;
            }
            if (!typeMap[type]) { // in case we dont know the type
                return data;
            }

            // Get the actual type of this object
            type = this.findCorrectType(data, type);

            // get the map for the correct type.
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            let instance: {[index: string]: any} = {};
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.baseName] = ObjectSerializer.serialize(data[attributeType.name], attributeType.type);
            }
            return instance;
        }
    }

    public static deserialize(data: any, type: string) {
        // polymorphism may change the actual type.
        type = ObjectSerializer.findCorrectType(data, type);
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.deserialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return new Date(data);
        } else {
            if (enumsMap[type]) {// is Enum
                return data;
            }

            if (!typeMap[type]) { // dont know the type
                return data;
            }
            let instance = new typeMap[type]();
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.name] = ObjectSerializer.deserialize(data[attributeType.baseName], attributeType.type);
            }
            return instance;
        }
    }
}

export interface Authentication {
    /**
    * Apply authentication settings to header and query params.
    */
    applyToRequest(requestOptions: localVarRequest.Options): Promise<void> | void;
}

export class HttpBasicAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        requestOptions.auth = {
            username: this.username, password: this.password
        }
    }
}

export class HttpBearerAuth implements Authentication {
    public accessToken: string | (() => string) = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            const accessToken = typeof this.accessToken === 'function'
                            ? this.accessToken()
                            : this.accessToken;
            requestOptions.headers["Authorization"] = "Bearer " + accessToken;
        }
    }
}

export class ApiKeyAuth implements Authentication {
    public apiKey: string = '';

    constructor(private location: string, private paramName: string) {
    }

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (this.location == "query") {
            (<any>requestOptions.qs)[this.paramName] = this.apiKey;
        } else if (this.location == "header" && requestOptions && requestOptions.headers) {
            requestOptions.headers[this.paramName] = this.apiKey;
        } else if (this.location == 'cookie' && requestOptions && requestOptions.headers) {
            if (requestOptions.headers['Cookie']) {
                requestOptions.headers['Cookie'] += '; ' + this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
            else {
                requestOptions.headers['Cookie'] = this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
        }
    }
}

export class OAuth implements Authentication {
    public accessToken: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            requestOptions.headers["Authorization"] = "Bearer " + this.accessToken;
        }
    }
}

export class VoidAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(_: localVarRequest.Options): void {
        // Do nothing
    }
}

export type Interceptor = (requestOptions: localVarRequest.Options) => (Promise<void> | void);
