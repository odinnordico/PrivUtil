import { createChannel, createClient } from 'nice-grpc-web';
import { PrivUtilServiceDefinition } from '../proto/proto/privutil';

const backendUrl = import.meta.env.VITE_API_URL || window.location.origin;
const channel = createChannel(backendUrl); 
export const client = createClient(PrivUtilServiceDefinition, channel);
