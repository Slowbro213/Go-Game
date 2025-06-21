// src/lib.rs

use wasm_bindgen::prelude::*;
use js_sys::{Uint8Array, Uint32Array, Float32Array};
use std::convert::TryInto;

#[wasm_bindgen]
pub fn decode(data: &[u8]) -> Result<JsValue, JsValue> {
    // total buffer length
    let total = data.len();
    let mut offset = 0;

    // ——— 1) read msg_type length (u32 LE) —————————————
    if total < 4 {
        return Err(JsValue::from_str("Buffer too small for msg_type length"));
    }
    let type_len = u32::from_le_bytes(data[0..4].try_into().unwrap()) as usize;
    offset += 4;

    // ——— 2) read msg_type string ———————————————————————
    if offset + type_len > total {
        return Err(JsValue::from_str("msg_type length exceeds buffer"));
    }
    let msg_type = std::str::from_utf8(&data[offset..offset + type_len])
        .map_err(|e| JsValue::from_str(&format!("Invalid UTF-8: {}", e)))?
        .to_string();
    offset += type_len;

    // ——— 3) parse objects into three primitive Vecs ————————
    // you always used "character".len() == 9
    const OBJ_TYPE_LEN: usize = 9;
    let mut ids = Vec::new();
    let mut xs  = Vec::new();
    let mut ys  = Vec::new();

    // each object in your old format was: 4(id)+9(type)+4(x)+4(y) bytes
    while offset + 4 + OBJ_TYPE_LEN + 8 <= total {
        // read id
        let id = u32::from_le_bytes(data[offset..offset+4].try_into().unwrap());
        offset += 4;

        // skip the 9-byte "character"
        offset += OBJ_TYPE_LEN;

        // read x,y
        let x = f32::from_le_bytes(data[offset..offset+4].try_into().unwrap());
        offset += 4;
        let y = f32::from_le_bytes(data[offset..offset+4].try_into().unwrap());
        offset += 4;

        ids.push(id);
        xs.push(x);
        ys.push(y);
    }

    // ——— 4) build JS TypedArrays in one go each ——————————
    let ids_array = Uint32Array::from(ids.as_slice());
    let xs_array  = Float32Array::from(xs.as_slice());
    let ys_array  = Float32Array::from(ys.as_slice());

    // ——— 5) package up a JS object: { type, ids, xs, ys } ————
    let result = js_sys::Object::new();
    js_sys::Reflect::set(&result, &"type".into(), &msg_type.into())?;
    js_sys::Reflect::set(&result, &"ids".into(), &JsValue::from(ids_array))?;
    js_sys::Reflect::set(&result, &"xs".into(), &JsValue::from(xs_array))?;
    js_sys::Reflect::set(&result, &"ys".into(), &JsValue::from(ys_array))?;

    Ok(JsValue::from(result))
}
