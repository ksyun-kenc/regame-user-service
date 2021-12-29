/*
 * Copyright 2021-present Ksyun
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const jayson = require('jayson')

const server = jayson.server({
  login: function (args, callback) {
    if (args.version !== 0) {
      callback({code: 1, message: 'Unsupported version!'})
      return
    }

    if (args.type === 0) {
      if (args.username === 'UMU' && args.data === '123456') {
        callback(null, {interval: 5, session_id: '64162c06-3770-44ef-9c85-19d640ba0155'})
        console.log('login', args)
        return
      }
    } else if (args.type === 1) {
      if (args.username === 'UMU' && args.data === '207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb') {
        callback(null, {interval: 5, session_id: '64162c06-3770-44ef-9c85-19d640ba0155'})
        return
      }
    } else if (args.type === 2) {
      if (args.username === 'UMU' && args.data === '64162c06-3770-44ef-9c85-19d640ba0155') {
        callback(null, {interval: 5, session_id: '64162c06-3770-44ef-9c85-19d640ba0155'})
        return
      }
    } else {
      callback({code: 2, message: 'Invalid type!'})
    }

    callback({code: 3, message: 'Failed!'})
  },
  keepalive: function (args, callback) {
    callback(null, {interval: 5, session_id: '64162c06-3770-44ef-9c85-19d640ba0155'})
    console.log('keepalive', args)
  },
  logout: function (args, callback) {
    callback(null)
    console.log('logout', args)
  }
})

server.http().listen(8545, '127.0.0.1')
