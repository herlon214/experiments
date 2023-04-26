# Image Search Experiment
Dataset: https://huggingface.co/datasets/sasha/dog-food

Goal: extract dominant colors from a dataset and index the images with their
respective dominant colors, then perform color quantization and also index them to
verify group sizes.


### Results
| Type               | Group 1 | Group 2 | Group 3 | Group 4 | Group 5 | Group 6 | Group 7 |
| ------------------ | ------- | ------- | ------- | ------- | ------- | ------- | ------- |
| Quantization 1bit  | 93      | 85      | 62      | 48      | 2       | 2       | 1       |
| Quantization 2bits | 68      | 51      | 45      | 44      | 39      | 33      | 28      |
| Quantization 3bits | 41      | 25      | 21      | 21      | 17      | 17      | 16      |
| Quantization 4bits | 25      | 11      | 5       | 4       | 4       | 3       | 3       |
| Quantization 5bits | 17      | 5       | 4       | 4       | 3       | 2       | 2       |
| Quantization 6bits | 7       | 5       | 4       | 3       | 2       | 2       | 2       |
| Quantization 7bits | 5       | 2       | 2       | 2       | 2       | 2       | 2       |
| No quantization    | 2       | 2       | 2       | 1       | 1       | 1       |         |


(Blocks follow the same group order, 1, 2, 3...)


1 bit:

![Block 1](./blocks_1bits.png)

2 bits:

![Block 2](./blocks_2bits.png)

3 bits:

![Block 3](./blocks_3bits.png)

4 bits:

![Block 4](./blocks_4bits.png)

5 bits:

![Block 5](./blocks_5bits.png)

6 bits:

![Block 6](./blocks_6bits.png)

7 bits:

![Block 7](./blocks_7bits.png)


No quantization:

![No quantization](./blocks_8bits.png)