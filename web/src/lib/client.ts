import { createChannel, createClient } from 'nice-grpc-web';
import { PrivUtilServiceDefinition } from '../proto/proto/privutil';

const channel = createChannel('http://localhost:8080'); // Helper to configure env later
export const client = createClient(PrivUtilServiceDefinition, channel);
