mod triangle;
mod decoder;

use wasm_bindgen::JsValue;
use wasm_bindgen::prelude::wasm_bindgen;

#[wasm_bindgen(start)]
pub fn start() -> Result< (), JsValue> {
    triangle::render_triangle()
}
