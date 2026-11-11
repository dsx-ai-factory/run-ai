import * as http from 'http';

export class HttpError extends Error {
    constructor (public response: http.IncomingMessage, public body: any, public statusCode?: number, requestOptions?: any) {
        super(`HTTP request failed with the following statusCode ${statusCode},\n Request: ${JSON.stringify(requestOptions)},\n Response: ${JSON.stringify(body)}`);
        this.name = 'HttpError';
    }
}
