#include<stdio.h>
#include<float.h>
#include<limits.h>

int main() {
        printf("float 存储最大字节数: %lu \n", sizeof(float));
        printf("float 最大值: %E\n", FLT_MIN);
        printf("float 最大值: %E\n", FLT_MAX);
        int a = 0;
        for (int i = 0; i < 100; i++) {
                a+=i;
        }
        printf("\n", a);
        return 0;
}