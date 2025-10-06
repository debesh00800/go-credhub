class Solution {
public:
    long long bowlSubarrays(vector<int>& nums) {
        stack<int> s;
        int n=nums.size();
        vector<int> right(n,n);
        for(int i=n-1;i>=0;i--){
            int curr=nums[i];
            while(!s.empty() && curr>nums[s.top()]){
                s.pop();
            }
            if(!s.empty()){
                right[i]=s.top();
            }
            s.push(i);
        }
        stack<int> s1;
        vector<int> left(n,-1);
        for(int i=0;i<n;i++){
            int curr=nums[i];
            while(!s1.empty() && curr>nums[s1.top()]){
                s1.pop();
            }
            if(!s1.empty()){
                left[i]=s1.top();
            }
            s1.push(i);
        }

        long long ans=0;
        for(int i=0;i<n;i++){
            if(right[i]!=n && right[i]-i+1>=3){ans++;}
            if(left[i]!=-1 && i-left[i]+1>=3){ans++;}
        }
        return ans;
            
    }
};
