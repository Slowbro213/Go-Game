// src/lib.rs
use wasm_bindgen::prelude::*;
use js_sys::{Uint32Array, Float32Array, Uint8Array, Object,Reflect};
use std::cell::UnsafeCell;


fn read_u32(buf: &[u8], off: &mut usize) -> Result<u32, JsValue> {
    if *off + 4 > buf.len() {
        return Err(JsValue::from_str(&format!("oob u32 at {}", *off)));
    }
    let v = u32::from_le_bytes(buf[*off..*off + 4].try_into().unwrap());
    *off += 4;
    Ok(v)
}

fn read_f32(buf: &[u8], off: &mut usize) -> Result<f32, JsValue> {
    if *off + 4 > buf.len() {
        return Err(JsValue::from_str(&format!("oob f32 at {}", *off)));
    }
    let v = f32::from_le_bytes(buf[*off..*off + 4].try_into().unwrap());
    *off += 4;
    Ok(v)
}

const MAX_OBJECTS: usize = 500000; // Adjust based on your max expected objects.

thread_local! {
    static OUTPUT: UnsafeCell<DecodeOutput> = UnsafeCell::new(DecodeOutput::new());
}

// Pre-allocated output buffers
struct DecodeOutput {
    ids: Vec<u32>,
    xs: Vec<f32>,
    ys: Vec<f32>,
    types: Vec<u8>,
    msg_type: String,
}

impl DecodeOutput {
    fn new() -> Self {
        Self {
            ids: Vec::with_capacity(MAX_OBJECTS),
            xs: Vec::with_capacity(MAX_OBJECTS),
            ys: Vec::with_capacity(MAX_OBJECTS),
            types: Vec::with_capacity(MAX_OBJECTS),
            msg_type: String::with_capacity(32),
        }
    }

    // Reset buffers (reuse allocated memory)
    fn reset(&mut self) {
        self.ids.clear();
        self.xs.clear();
        self.ys.clear();
        self.types.clear();
        self.msg_type.clear();
    }
}

#[wasm_bindgen]
pub fn decode(buf: &[u8]) -> Result<JsValue, JsValue> {
    let mut off = 0;
    let total = buf.len();

    OUTPUT.with(|output| {
        let output = unsafe { &mut *output.get() };
        output.reset();

        // 1) Parse message type (zero-copy if possible)
        let type_len = read_u32(buf, &mut off)? as usize;
        if off + type_len > total {
            return Err(JsValue::from_str("oob msg type"));
        }
        output.msg_type = std::str::from_utf8(&buf[off..off + type_len])
            .map_err(|_| JsValue::from_str("bad utf8 in msg type"))?
            .to_string();
        off += type_len;

        // 2) Parse objects (zero-copy into pre-allocated buffers)
        while off < total {
            output.ids.push(read_u32(buf, &mut off)?);
            output.types.push(buf[off]);
            off += 1;
            output.xs.push(read_f32(buf, &mut off)?);
            output.ys.push(read_f32(buf, &mut off)?);
        }

        // 3) Expose views into Wasm memory (zero-copy)
        let out = Object::new();
        unsafe{
        Reflect::set(&out, &"type".into(), &JsValue::from_str(&output.msg_type))?;
        Reflect::set(&out, &"ids".into(), &Uint32Array::view(&output.ids).into())?;
        Reflect::set(&out, &"xs".into(), &Float32Array::view(&output.xs).into())?;
        Reflect::set(&out, &"ys".into(), &Float32Array::view(&output.ys).into())?;
        Reflect::set(&out, &"typeCodes".into(), &Uint8Array::view(&output.types).into())?;
        }
        Ok(out.into())
    })
}

// Keep your existing read_u32/read_f32 helpers.
