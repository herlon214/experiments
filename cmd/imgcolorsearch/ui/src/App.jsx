import { useState, useEffect } from "react";
import algoliasearch from "algoliasearch";
import { HexColorPicker } from "react-colorful";
import reactLogo from "./assets/react.svg";
import viteLogo from "/vite.svg";
import { useDebounce } from "usehooks-ts";
import "./App.css";

const API_KEY = import.meta.env.VITE_API_KEY;
const APP_ID = import.meta.env.VITE_APP_ID;
const client = algoliasearch(APP_ID, API_KEY);
const index = client.initIndex("exp_image_search");

function hexToRgb(hex) {
  var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
}

function quantizeColor(col, bits) {
  const maxValue = Math.pow(2, bits);
  const colorStep = 256 / maxValue;

  const rQuantized = Math.floor(col.r / colorStep) * colorStep;
  const gQuantized = Math.floor(col.g / colorStep) * colorStep;
  const bQuantized = Math.floor(col.b / colorStep) * colorStep;

  return { r: rQuantized, g: gQuantized, b: bQuantized, a: col.a };
}

function componentToHex(c) {
  var hex = c.toString(16);
  return hex.length == 1 ? "0" + hex : hex;
}

function rgbToHex(r, g, b) {
  return "#" + componentToHex(r) + componentToHex(g) + componentToHex(b);
}

function App() {
  const [color, setColor] = useState("#ffffff", 500);
  const [count, setCount] = useState(0);
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [quantizedValues, setQuantizedValues] = useState([]);
  const debouncedValue = useDebounce(color, 500);
  const rgb = hexToRgb(debouncedValue);

  useEffect(() => {
    let values = [];
    for (let i = 1; i < 9; i++) {
      const val = quantizeColor(rgb, i);
      values.push(rgbToHex(val.r, val.g, val.b));
    }

    setQuantizedValues(values);

    index
      .search("", {
        facetFilters: [
          [
            "quantizedColorsHex.8:" + quantizedValues[7],
            "quantizedColorsHex.7:" + quantizedValues[6],
            "quantizedColorsHex.6:" + quantizedValues[5],
            "quantizedColorsHex.5:" + quantizedValues[4],
            "quantizedColorsHex.4:" + quantizedValues[3],
            "quantizedColorsHex.3:" + quantizedValues[2],
            "quantizedColorsHex.2:" + quantizedValues[1],
          ],
        ],
        optionalFilters: [["quantizedColorsHex.1:" + quantizedValues[0]]],
      })
      .then(({ hits, nbHits }) => {
        setItems(hits);
        setTotal(nbHits);
      })
      .catch(console.log);
  }, [debouncedValue]);

  return (
    <>
      <h1 className="text-5xl font-bold">Image Search</h1>
      <div className="grid grid-cols-8">
        <div className="col-span-2">
          <HexColorPicker className="mt-10" color={color} onChange={setColor} />
          <div
            className="mt-5"
            style={{
              backgroundColor: color,
              height: "100px",
              width: "200px",
              display: "block",
            }}
          ></div>
          <div className="flex flex-col mt-10">
            <br />
            {quantizedValues.map((item, i) => (
              <div
                key={i}
                style={{
                  backgroundColor: item,
                  height: "50px",
                  width: "50px",
                }}
              ></div>
            ))}

            <p className="pt-10 mr-5">
              RGB({rgb.r}, {rgb.g}, {rgb.b})<br />
              Quantized Hex: <br />
              {quantizedValues.map((item, i) => (
                <>
                  <b key={i}>
                    {i + 1} bits - {item}
                  </b>{" "}
                  <br />
                </>
              ))}
            </p>
          </div>
        </div>
        <div className="col-span-6 mt-10">
          <strong>Found {total} image(s).</strong>
          <div className="mt-20 flex flex-row flex-wrap">
            {items.map((item) => (
              <img
                key={item.objectID}
                src={item.imageURL}
                style={{ maxHeight: "200px" }}
              />
            ))}
          </div>
        </div>
      </div>
    </>
  );
}

export default App;
