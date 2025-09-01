import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
    vus: 20,
    duration: '2m',
};

export default function () {
  http.get('http://localhost:4200/')
  
  // Add random sleep between 0.1 and 0.5 seconds
  sleep(Math.random() * 0.4 + 0.1)
}
