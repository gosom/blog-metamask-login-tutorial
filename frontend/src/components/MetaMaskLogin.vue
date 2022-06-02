<template >
  <button @click="login">
    <img src="../assets/metamask-logo.svg"  
    width="60" height="60"
    @click="login"
    />
    <span>Login with Metamask</span>
  </button>
</template>


<script setup>
  import { Buffer } from 'buffer'
  const ethereum = window.ethereum
  var account = null
  var address = null
  var token = null

  console.log(token)

  async function login() {
    if (address === null || account === null){
      const accounts = await ethereum.request({ method: 'eth_requestAccounts' })
      account = accounts[0]
      address = ethereum.selectedAddress
    }
    const [status_code, nonce] = await get_nonce()
    if (status_code === 404) {
      const registered = await register()
      if (!registered) {
        return
      }
      await login()
      return
    }else if (status_code != 200) {
      return
    }

    const signature = await sign(nonce)
    token = await perform_signin(signature, nonce)
  }

  async function get_nonce() {
    const reqOpts = {
      method: "GET",
      headers: {"Content-Type": "application/json"},
    }
    const response = await fetch("http://localhost:8001/users/"+address+"/nonce", reqOpts)
    if (response.status === 200) {
      const data = await response.json()
      const nonce = data.Nonce
      return [200, nonce]
    }
    return [response.status, ""]
  }

  async function register() {
    const reqOpts = {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ 
        address: address, 
      })
    }
    const response = await fetch("http://localhost:8001/register", reqOpts)
    if (response.status === 201) {
      return true
    }
    return false
  }

  async function sign(nonce) {
    const buff = Buffer.from(nonce, "utf-8");
    const signature = await ethereum.request({
      method: "personal_sign",
      params: [ buff.toString("hex"), account],
    })
    return signature
  }

  async function perform_signin(sig, nonce) {
    const reqOpts = {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ 
        address: address, 
        nonce: nonce,
        sig: sig,
        })
    }
    const response = await fetch("http://localhost:8001/signin", reqOpts)
    if (response.status === 200) {
      const data = await response.json()
      return data
    }
    return null
  }

</script>
