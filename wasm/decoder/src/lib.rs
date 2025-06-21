// src/lib.rs
use wasm_bindgen::prelude::*;
use js_sys::{Uint32Array, Float32Array, Uint8Array, Object, Reflect};

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

/// decode to { type, ids, xs, ys, typeCodes }
#[wasm_bindgen]
pub fn decode(buf: &[u8]) -> Result<JsValue, JsValue> {
    let mut off = 0;
    let total = buf.len();

    // 1. top-level message type
    let type_len = read_u32(buf, &mut off)? as usize;
    if off + type_len > total {
        return Err(JsValue::from_str("oob msg type"));
    }
    let msg_type = std::str::from_utf8(&buf[off..off + type_len])
        .map_err(|_| JsValue::from_str("bad utf8 in msg type"))?;
    off += type_len;

    // 2. object loop
    let mut ids = Vec::new();
    let mut xs = Vec::new();
    let mut ys = Vec::new();
    let mut type_codes = Vec::new();

    while off < total {
        // ID
        ids.push(read_u32(buf, &mut off)?);

        type_codes.push(buf[off]);
        log(type_codes.)
        off += 1;

        // X, Y
        xs.push(read_f32(buf, &mut off)?);
        ys.push(read_f32(buf, &mut off)?);
    }

    // 3. views
    let ids_arr = unsafe { Uint32Array::view(&ids) };
    let xs_arr = unsafe { Float32Array::view(&xs) };
    let ys_arr = unsafe { Float32Array::view(&ys) };
    let types_arr = unsafe { Uint8Array::view(&type_codes) };

    // 4. return object
    let out = Object::new();
    Reflect::set(&out, &"type".into(), &msg_type.into())?;
    Reflect::set(&out, &"ids".into(), &ids_arr.into())?;
    Reflect::set(&out, &"xs".into(), &xs_arr.into())?;
    Reflect::set(&out, &"ys".into(), &ys_arr.into())?;
    Reflect::set(&out, &"typeCodes".into(), &types_arr.into())?;

    Ok(out.into())
}
