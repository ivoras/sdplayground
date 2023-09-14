#!/usr/bin/python3
import sys,os
import requests

data = {
  "key": "",
  "model_id": "midjourney",
  "prompt": "ultra realistic close up portrait ((beautiful pale cyberpunk female with heavy black eyeliner)), blue eyes, shaved side haircut, hyper detail, cinematic lighting, magic neon, dark red city, Canon EOS R3, nikon, f/1.4, ISO 200, 1/160s, 8K, RAW, unedited, symmetrical balance, in-frame, 8K",
  "negative_prompt": "painting, extra fingers, mutated hands, poorly drawn hands, poorly drawn face, deformed, ugly, blurry, bad anatomy, bad proportions, extra limbs, cloned face, skinny, glitchy, double torso, extra arms, extra hands, mangled fingers, missing lips, ugly face, distorted face, extra legs, anime",
  "width": "1024",
  "height": "1024",
  "samples": "1",
  "num_inference_steps": "30",
  "guidance_scale": 7.5,
}

resp = requests.post('https://stablediffusionapi.com/api/v4/dreambooth', json=data)

print(resp.text)

